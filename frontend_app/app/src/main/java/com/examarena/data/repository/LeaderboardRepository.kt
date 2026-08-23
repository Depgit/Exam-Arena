package com.examarena.data.repository

import com.examarena.core.common.Resource
import com.examarena.core.network.ApiClient
import com.examarena.data.model.LeaderboardEntryDto
import com.examarena.data.model.LeaderboardResponseDto
import kotlinx.serialization.json.decodeFromJsonElement
import kotlinx.serialization.json.jsonArray
import kotlinx.serialization.json.jsonObject

class LeaderboardRepository(private val apiClient: ApiClient) {

    suspend fun getLeaderboard(
        category: String,
        limit: Int = 50,
        offset: Int = 0
    ): Resource<List<LeaderboardEntryDto>> {
        return apiClient.get(
            "/api/v1/leaderboard/$category?limit=$limit&offset=$offset",
            requireAuth = false
        ) { body ->
            try {
                val element = apiClient.json.parseToJsonElement(body)
                if (element is kotlinx.serialization.json.JsonObject && element.containsKey("data")) {
                    val dataArray = element.jsonObject["data"]?.jsonArray
                    if (dataArray != null) {
                        apiClient.json.decodeFromJsonElement<List<LeaderboardEntryDto>>(dataArray)
                    } else emptyList()
                } else if (element is kotlinx.serialization.json.JsonArray) {
                    apiClient.json.decodeFromJsonElement<List<LeaderboardEntryDto>>(element)
                } else {
                    emptyList()
                }
            } catch (e: Exception) {
                emptyList()
            }
        }
    }
}
