package com.examarena.data.repository

import com.examarena.core.common.Resource
import com.examarena.core.network.ApiClient
import com.examarena.data.model.PracticeSessionResponseDto
import com.examarena.data.model.StartPracticeRequestDto
import com.examarena.data.model.SubmitPracticeAnswerRequestDto
import com.examarena.data.model.SubmitPracticeAnswerResponseDto
import kotlinx.serialization.encodeToString

class PracticeRepository(private val apiClient: ApiClient) {

    suspend fun startPractice(
        examCategoryId: String,
        difficulty: String = "mixed",
        questionCount: Int = 5
    ): Resource<PracticeSessionResponseDto> {
        val requestDto = StartPracticeRequestDto(
            examCategoryId = examCategoryId,
            difficulty = difficulty,
            questionCount = questionCount
        )
        val bodyJson = apiClient.json.encodeToString(requestDto)

        return apiClient.post("/api/v1/practice/start", bodyJson, requireAuth = true) { body ->
            apiClient.json.decodeFromString<PracticeSessionResponseDto>(body)
        }
    }

    suspend fun submitAnswer(
        sessionId: String,
        questionId: String,
        selectedOptionId: String,
        timeTakenMs: Int
    ): Resource<SubmitPracticeAnswerResponseDto> {
        val requestDto = SubmitPracticeAnswerRequestDto(
            questionId = questionId,
            selectedOptionId = selectedOptionId,
            timeTakenMs = timeTakenMs
        )
        val bodyJson = apiClient.json.encodeToString(requestDto)

        return apiClient.post("/api/v1/practice/$sessionId/answer", bodyJson, requireAuth = true) { body ->
            apiClient.json.decodeFromString<SubmitPracticeAnswerResponseDto>(body)
        }
    }

    suspend fun endSession(sessionId: String): Resource<String> {
        return apiClient.post("/api/v1/practice/$sessionId/end", "{}", requireAuth = true) { body ->
            body
        }
    }
}
