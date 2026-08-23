package com.examarena.data.repository

import com.examarena.core.common.Resource
import com.examarena.core.network.ApiClient
import com.examarena.core.websocket.ExamArenaWebSocket
import com.examarena.core.websocket.WsConnectionState
import com.examarena.core.websocket.WsEvent
import com.examarena.data.model.CreateFriendMatchRequestDto
import com.examarena.data.model.FriendMatchResponseDto
import com.examarena.data.model.JoinFriendMatchRequestDto
import com.examarena.data.model.JoinQueueRequestDto
import com.examarena.data.model.MatchDetailsResponseDto
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.serialization.encodeToString

class MatchRepository(
    private val apiClient: ApiClient,
    private val webSocket: ExamArenaWebSocket
) {
    val wsConnectionState: StateFlow<WsConnectionState> = webSocket.connectionState
    val wsEvents: SharedFlow<WsEvent> = webSocket.events

    fun connectWebSocket() {
        webSocket.connect()
    }

    fun disconnectWebSocket() {
        webSocket.disconnect()
    }

    suspend fun joinQueue(examCategoryId: String, matchType: String = "ranked"): Resource<String> {
        val requestDto = JoinQueueRequestDto(examCategoryId, matchType)
        val bodyJson = apiClient.json.encodeToString(requestDto)

        return apiClient.post("/api/v1/matches/queue", bodyJson, requireAuth = true) { body ->
            body
        }
    }

    suspend fun leaveQueue(): Resource<String> {
        return apiClient.delete("/api/v1/matches/queue", requireAuth = true) { body ->
            body
        }
    }

    suspend fun createFriendMatch(examCategoryId: String): Resource<FriendMatchResponseDto> {
        val requestDto = CreateFriendMatchRequestDto(examCategoryId)
        val bodyJson = apiClient.json.encodeToString(requestDto)

        return apiClient.post("/api/v1/matches/friend", bodyJson, requireAuth = true) { body ->
            apiClient.json.decodeFromString<FriendMatchResponseDto>(body)
        }
    }

    suspend fun joinFriendMatch(roomCode: String): Resource<FriendMatchResponseDto> {
        val requestDto = JoinFriendMatchRequestDto(roomCode)
        val bodyJson = apiClient.json.encodeToString(requestDto)

        return apiClient.post("/api/v1/matches/friend/join", bodyJson, requireAuth = true) { body ->
            apiClient.json.decodeFromString<FriendMatchResponseDto>(body)
        }
    }

    suspend fun getMatchDetails(matchId: String): Resource<MatchDetailsResponseDto> {
        return apiClient.get("/api/v1/matches/$matchId", requireAuth = true) { body ->
            apiClient.json.decodeFromString<MatchDetailsResponseDto>(body)
        }
    }

    fun submitAnswer(matchId: String, questionId: String, optionId: String, timeTakenMs: Int) {
        webSocket.submitAnswer(matchId, questionId, optionId, timeTakenMs)
    }
}
