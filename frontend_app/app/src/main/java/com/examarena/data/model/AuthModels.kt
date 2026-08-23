package com.examarena.data.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class RegisterRequestDto(
    val username: String,
    val email: String,
    val password: String
)

@Serializable
data class LoginRequestDto(
    val login: String,
    val password: String
)

@Serializable
data class FirebaseLoginRequestDto(
    @SerialName("id_token") val idToken: String,
    val email: String,
    val username: String? = null
)

@Serializable
data class AuthResponseDto(
    val token: String,
    val user: UserDto
)

@Serializable
data class UserDto(
    val id: String,
    val username: String,
    val email: String? = null,
    @SerialName("display_name") val displayName: String? = null,
    @SerialName("avatar_url") val avatarUrl: String? = null,
    @SerialName("country_code") val countryCode: String? = null,
    @SerialName("preferred_language") val preferredLanguage: String = "en",
    val role: String = "user",
    val status: String = "active"
)

@Serializable
data class UserProfileResponseDto(
    val user: UserDto,
    val ratings: List<UserRatingDto> = emptyList()
)

@Serializable
data class UserRatingDto(
    @SerialName("user_id") val userId: String,
    @SerialName("exam_category_id") val examCategoryId: String,
    val rating: Int = 1200,
    @SerialName("matches_played") val matchesPlayed: Int = 0
)

@Serializable
data class UserStatisticsDto(
    @SerialName("user_id") val userId: String,
    @SerialName("exam_category_id") val examCategoryId: String,
    @SerialName("total_matches") val totalMatches: Int = 0,
    val wins: Int = 0,
    val losses: Int = 0,
    val draws: Int = 0,
    @SerialName("current_win_streak") val currentWinStreak: Int = 0,
    @SerialName("longest_win_streak") val longestWinStreak: Int = 0,
    @SerialName("total_questions_solved") val totalQuestionsSolved: Int = 0,
    @SerialName("total_practice_sessions") val totalPracticeSessions: Int = 0,
    @SerialName("overall_accuracy") val overallAccuracy: Double = 0.0,
    @SerialName("avg_solving_time_ms") val avgSolvingTimeMs: Int? = null
)
