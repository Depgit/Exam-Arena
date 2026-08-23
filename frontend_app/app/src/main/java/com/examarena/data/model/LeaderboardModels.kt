package com.examarena.data.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class LeaderboardResponseDto(
    val data: List<LeaderboardEntryDto> = emptyList(),
    val meta: LeaderboardMetaDto? = null
)

@Serializable
data class LeaderboardEntryDto(
    val rank: Int,
    @SerialName("user_id") val userId: String,
    val username: String,
    @SerialName("display_name") val displayName: String? = null,
    @SerialName("avatar_url") val avatarUrl: String? = null,
    val rating: Int = 1200,
    @SerialName("matches_played") val matchesPlayed: Int = 0
)

@Serializable
data class LeaderboardMetaDto(
    val category: String? = null,
    val limit: Int = 50,
    val offset: Int = 0
)

enum class ExamCategory(
    val id: String,
    val code: String,
    val title: String,
    val description: String,
    val iconEmoji: String
) {
    SSCCGL("ssc-cgl", "SSC", "SSC CGL", "Staff Selection Commission exams", "🏛️"),
    BANKING("banking", "BANK", "Banking / IBPS", "IBPS, SBI and other bank exams", "🏦"),
    RAILWAYS("railways", "RRB", "Railways", "RRB & Railway recruitment", "🚆"),
    UPSC("upsc", "UPSC", "UPSC Prelims", "Civil Services Prelims", "📜"),
    CAT("cat", "CAT", "CAT / MBA", "Common Admission & MBA tests", "📈"),
    STATEPSC("state-psc", "PSC", "State PSC", "State Public Service tests", "🗺️");

    companion object {
        fun fromId(id: String): ExamCategory {
            return entries.find { it.id.equals(id, ignoreCase = true) } ?: SSCCGL
        }
    }
}
