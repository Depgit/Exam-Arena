package com.examarena.core.websocket

import android.util.Log
import com.examarena.core.datastore.SessionManager
import com.examarena.core.network.ApiClient
import com.examarena.data.model.MatchPlayerInfo
import com.examarena.data.model.MatchResultEntry
import com.examarena.data.model.QuestionForPlayer
import com.examarena.data.model.ScoreboardEntry
import com.examarena.data.model.SubmitAnswerWsPayload
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.decodeFromJsonElement
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import java.util.concurrent.TimeUnit

class ExamArenaWebSocket(
    private val sessionManager: SessionManager,
    private val apiClient: ApiClient
) {
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    private val _connectionState = MutableStateFlow(WsConnectionState.Disconnected)
    val connectionState: StateFlow<WsConnectionState> = _connectionState.asStateFlow()

    private val _events = MutableSharedFlow<WsEvent>(extraBufferCapacity = 64)
    val events: SharedFlow<WsEvent> = _events.asSharedFlow()

    private var webSocket: WebSocket? = null
    private var heartbeatJob: Job? = null
    private var reconnectJob: Job? = null
    private var isManualDisconnect = false

    private val okHttpClient = OkHttpClient.Builder()
        .readTimeout(0, TimeUnit.MILLISECONDS) // infinite for WS
        .pingInterval(25, TimeUnit.SECONDS)
        .build()

    fun connect() {
        isManualDisconnect = false
        if (_connectionState.value == WsConnectionState.Connected ||
            _connectionState.value == WsConnectionState.Connecting
        ) {
            return
        }

        scope.launch {
            val token = sessionManager.getAuthToken()
            if (token.isNullOrBlank()) {
                Log.w("ExamArenaWS", "Cannot connect WebSocket: No auth token found.")
                _connectionState.value = WsConnectionState.Disconnected
                return@launch
            }

            val baseWsUrl = sessionManager.getWsUrl()
            val wsUrl = "$baseWsUrl?token=$token"

            _connectionState.value = WsConnectionState.Connecting

            val request = Request.Builder()
                .url(wsUrl)
                .build()

            webSocket = okHttpClient.newWebSocket(request, createListener())
        }
    }

    fun disconnect() {
        isManualDisconnect = true
        heartbeatJob?.cancel()
        reconnectJob?.cancel()
        webSocket?.close(1000, "Client disconnect")
        webSocket = null
        _connectionState.value = WsConnectionState.Disconnected
    }

    fun submitAnswer(matchId: String, questionId: String, optionId: String, timeTakenMs: Int) {
        val payload = SubmitAnswerWsPayload(
            matchId = matchId,
            questionId = questionId,
            optionId = optionId,
            timeTakenMs = timeTakenMs
        )
        val jsonPayload = apiClient.json.encodeToJsonElement(SubmitAnswerWsPayload.serializer(), payload)
        val rawMessage = WsRawMessage(
            type = "submit_answer",
            payload = jsonPayload
        )
        val messageStr = apiClient.json.encodeToString(rawMessage)
        webSocket?.send(messageStr)
    }

    fun sendPing() {
        val message = WsRawMessage(type = "ping")
        val messageStr = apiClient.json.encodeToString(message)
        webSocket?.send(messageStr)
    }

    private fun startHeartbeat() {
        heartbeatJob?.cancel()
        heartbeatJob = scope.launch {
            while (isActive) {
                delay(20000)
                if (_connectionState.value == WsConnectionState.Connected) {
                    sendPing()
                }
            }
        }
    }

    private fun scheduleReconnect() {
        if (isManualDisconnect) return
        reconnectJob?.cancel()
        reconnectJob = scope.launch {
            _connectionState.value = WsConnectionState.Reconnecting
            delay(3000)
            if (!isManualDisconnect && _connectionState.value != WsConnectionState.Connected) {
                Log.i("ExamArenaWS", "Attempting automatic WebSocket reconnection...")
                connect()
            }
        }
    }

    private fun createListener(): WebSocketListener {
        return object : WebSocketListener() {
            override fun onOpen(webSocket: WebSocket, response: Response) {
                Log.i("ExamArenaWS", "WebSocket connected successfully")
                _connectionState.value = WsConnectionState.Connected
                startHeartbeat()
            }

            override fun onMessage(webSocket: WebSocket, text: String) {
                parseAndDispatchMessage(text)
            }

            override fun onClosing(webSocket: WebSocket, code: Int, reason: String) {
                _connectionState.value = WsConnectionState.Disconnected
            }

            override fun onClosed(webSocket: WebSocket, code: Int, reason: String) {
                _connectionState.value = WsConnectionState.Disconnected
                if (!isManualDisconnect) {
                    scheduleReconnect()
                }
            }

            override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
                Log.e("ExamArenaWS", "WebSocket failure: ${t.message}")
                _connectionState.value = WsConnectionState.Failed
                if (!isManualDisconnect) {
                    scheduleReconnect()
                }
            }
        }
    }

    private fun parseAndDispatchMessage(text: String) {
        scope.launch {
            try {
                val raw = apiClient.json.decodeFromString<WsRawMessage>(text)
                val payload = raw.payload

                val event: WsEvent = when (raw.type) {
                    "connected" -> {
                        val userId = payload?.jsonObject?.get("user_id")?.jsonPrimitive?.content ?: ""
                        val username = payload?.jsonObject?.get("username")?.jsonPrimitive?.content ?: ""
                        val message = payload?.jsonObject?.get("message")?.jsonPrimitive?.content ?: ""
                        WsEvent.Connected(userId, username, message)
                    }
                    "match_start" -> {
                        if (payload != null) {
                            val matchId = payload.jsonObject["match_id"]?.jsonPrimitive?.content ?: ""
                            val matchType = payload.jsonObject["match_type"]?.jsonPrimitive?.content ?: "ranked"
                            val timerSeconds = payload.jsonObject["timer_seconds"]?.jsonPrimitive?.content?.toIntOrNull() ?: 120
                            val questionsJson = payload.jsonObject["questions"]
                            val playersJson = payload.jsonObject["players"]

                            val questions: List<QuestionForPlayer> = if (questionsJson != null) {
                                apiClient.json.decodeFromJsonElement(questionsJson)
                            } else emptyList()

                            val players: List<MatchPlayerInfo> = if (playersJson != null) {
                                apiClient.json.decodeFromJsonElement(playersJson)
                            } else emptyList()

                            WsEvent.MatchStart(matchId, matchType, timerSeconds, questions, players)
                        } else {
                            WsEvent.Raw(raw.type, text)
                        }
                    }
                    "score_update" -> {
                        if (payload != null) {
                            val matchId = payload.jsonObject["match_id"]?.jsonPrimitive?.content ?: ""
                            val userId = payload.jsonObject["user_id"]?.jsonPrimitive?.content ?: ""
                            val questionId = payload.jsonObject["question_id"]?.jsonPrimitive?.content ?: ""
                            val isCorrect = payload.jsonObject["is_correct"]?.jsonPrimitive?.content?.toBooleanStrictOrNull() ?: false
                            val pointsEarned = payload.jsonObject["points_earned"]?.jsonPrimitive?.content?.toIntOrNull() ?: 0
                            val scoreboardJson = payload.jsonObject["scoreboard"]

                            val scoreboard: List<ScoreboardEntry> = if (scoreboardJson != null) {
                                apiClient.json.decodeFromJsonElement(scoreboardJson)
                            } else emptyList()

                            WsEvent.ScoreUpdate(matchId, userId, questionId, isCorrect, pointsEarned, scoreboard)
                        } else {
                            WsEvent.Raw(raw.type, text)
                        }
                    }
                    "time_update" -> {
                        val matchId = payload?.jsonObject?.get("match_id")?.jsonPrimitive?.content ?: ""
                        val remaining = payload?.jsonObject?.get("remaining_seconds")?.jsonPrimitive?.content?.toIntOrNull() ?: 0
                        WsEvent.TimeUpdate(matchId, remaining)
                    }
                    "match_end" -> {
                        if (payload != null) {
                            val matchId = payload.jsonObject["match_id"]?.jsonPrimitive?.content ?: ""
                            val resultsJson = payload.jsonObject["results"]
                            val results: List<MatchResultEntry> = if (resultsJson != null) {
                                apiClient.json.decodeFromJsonElement(resultsJson)
                            } else emptyList()
                            WsEvent.MatchEnd(matchId, results)
                        } else {
                            WsEvent.Raw(raw.type, text)
                        }
                    }
                    "match_failed" -> {
                        val reason = payload?.jsonObject?.get("reason")?.jsonPrimitive?.content ?: "Match failed"
                        WsEvent.MatchFailed(reason)
                    }
                    "pong" -> WsEvent.Pong
                    "error" -> {
                        val message = payload?.jsonObject?.get("message")?.jsonPrimitive?.content ?: "Server error"
                        val matchId = payload?.jsonObject?.get("match_id")?.jsonPrimitive?.content
                        WsEvent.Error(message, matchId)
                    }
                    else -> WsEvent.Raw(raw.type, text)
                }

                _events.emit(event)
            } catch (e: Exception) {
                Log.e("ExamArenaWS", "Error parsing WebSocket message: $text", e)
                _events.emit(WsEvent.Raw("parse_error", text))
            }
        }
    }
}
