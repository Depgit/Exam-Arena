package com.examarena.core.websocket

import com.examarena.data.model.MatchPlayerInfo
import com.examarena.data.model.MatchResultEntry
import com.examarena.data.model.QuestionForPlayer
import com.examarena.data.model.ScoreboardEntry
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonElement

@Serializable
data class WsRawMessage(
    val type: String,
    val payload: JsonElement? = null
)

sealed interface WsEvent {
    data class Connected(val userId: String, val username: String, val message: String) : WsEvent
    data class MatchStart(
        val matchId: String,
        val matchType: String,
        val timerSeconds: Int,
        val questions: List<QuestionForPlayer>,
        val players: List<MatchPlayerInfo>
    ) : WsEvent
    data class ScoreUpdate(
        val matchId: String,
        val userId: String,
        val questionId: String,
        val isCorrect: Boolean,
        val pointsEarned: Int,
        val scoreboard: List<ScoreboardEntry>
    ) : WsEvent
    data class TimeUpdate(val matchId: String, val remainingSeconds: Int) : WsEvent
    data class MatchEnd(val matchId: String, val results: List<MatchResultEntry>) : WsEvent
    data class MatchFailed(val reason: String) : WsEvent
    data class Error(val message: String, val matchId: String? = null) : WsEvent
    data object Pong : WsEvent
    data class Raw(val type: String, val raw: String) : WsEvent
}

enum class WsConnectionState {
    Disconnected,
    Connecting,
    Connected,
    Reconnecting,
    Failed
}
