package com.examarena.data.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class JoinQueueRequestDto(
    @SerialName("exam_category_id") val examCategoryId: String,
    @SerialName("match_type") val matchType: String = "ranked"
)

@Serializable
data class CreateFriendMatchRequestDto(
    @SerialName("exam_category_id") val examCategoryId: String
)

@Serializable
data class JoinFriendMatchRequestDto(
    @SerialName("room_code") val roomCode: String
)

@Serializable
data class FriendMatchResponseDto(
    @SerialName("match_id") val matchId: String,
    @SerialName("room_code") val roomCode: String? = null,
    val status: String,
    val message: String? = null
)

@Serializable
data class MatchDto(
    val id: String,
    @SerialName("match_type") val matchType: String,
    val status: String,
    @SerialName("exam_category_id") val examCategoryId: String,
    @SerialName("room_code") val roomCode: String? = null,
    @SerialName("timer_seconds") val timerSeconds: Int = 120,
    @SerialName("max_players") val maxPlayers: Int = 2
)

@Serializable
data class MatchPlayerDto(
    val id: String? = null,
    @SerialName("match_id") val matchId: String,
    @SerialName("user_id") val userId: String,
    val username: String,
    val score: Int = 0,
    @SerialName("final_rank") val finalRank: Int? = null,
    @SerialName("rating_before") val ratingBefore: Int? = null,
    @SerialName("rating_after") val ratingAfter: Int? = null,
    @SerialName("rating_delta") val ratingDelta: Int? = null,
    @SerialName("connection_status") val connectionStatus: String = "connected"
)

@Serializable
data class MatchDetailsResponseDto(
    val match: MatchDto,
    val players: List<MatchPlayerDto> = emptyList(),
    val questions: List<QuestionForPlayer> = emptyList(),
    @SerialName("live_scores") val liveScores: List<ScoreboardEntry> = emptyList()
)

@Serializable
data class QuestionForPlayer(
    val id: String,
    @SerialName("question_type") val questionType: String = "mcq_single",
    val difficulty: String = "medium",
    val body: String,
    @SerialName("estimated_time_seconds") val estimatedTimeSeconds: Int = 30,
    val options: List<OptionForPlayer> = emptyList(),
    @SerialName("order_index") val orderIndex: Int = 1
)

@Serializable
data class OptionForPlayer(
    val id: String,
    @SerialName("option_text") val optionText: String,
    @SerialName("order_index") val orderIndex: Int = 1
)

@Serializable
data class MatchPlayerInfo(
    @SerialName("user_id") val userId: String,
    val username: String,
    val rating: Int? = null
)

@Serializable
data class ScoreboardEntry(
    @SerialName("user_id") val userId: String,
    val username: String,
    val score: Int,
    @SerialName("questions_answered") val questionsAnswered: Int = 0,
    val correct: Int = 0
)

@Serializable
data class MatchResultEntry(
    @SerialName("user_id") val userId: String,
    val username: String,
    val score: Int,
    val rank: Int,
    val correct: Int = 0,
    val total: Int = 0
)

@Serializable
data class SubmitAnswerWsPayload(
    @SerialName("match_id") val matchId: String,
    @SerialName("question_id") val questionId: String,
    @SerialName("option_id") val optionId: String,
    @SerialName("time_taken_ms") val timeTakenMs: Int
)
