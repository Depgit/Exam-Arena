package com.examarena.data.repository

import com.examarena.core.common.Resource
import com.examarena.core.network.ApiClient
import com.examarena.data.model.UserProfileResponseDto
import com.examarena.data.model.UserStatisticsDto

class UserRepository(private val apiClient: ApiClient) {

    suspend fun getUserProfile(userId: String): Resource<UserProfileResponseDto> {
        return apiClient.get("/api/v1/users/$userId", requireAuth = false) { body ->
            apiClient.json.decodeFromString<UserProfileResponseDto>(body)
        }
    }

    suspend fun getUserStatistics(userId: String): Resource<UserStatisticsDto> {
        return apiClient.get("/api/v1/users/$userId/stats", requireAuth = false) { body ->
            apiClient.json.decodeFromString<UserStatisticsDto>(body)
        }
    }
}
