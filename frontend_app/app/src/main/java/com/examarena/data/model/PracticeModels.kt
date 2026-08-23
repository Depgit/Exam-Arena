package com.examarena.data.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class StartPracticeRequestDto(
    @SerialName("exam_category_id") val examCategoryId: String,
    @SerialName("topic_id") val topicId: String? = null,
    val difficulty: String = "mixed",
    @SerialName("question_count") val questionCount: Int = 5
)

@Serializable
data class PracticeSessionResponseDto(
    val session: PracticeSessionDto,
    val questions: List<PracticeQuestionDto> = emptyList()
)

@Serializable
data class PracticeSessionDto(
    val id: String,
    @SerialName("user_id") val userId: String,
    @SerialName("exam_category_id") val examCategoryId: String,
    val difficulty: String? = null,
    @SerialName("question_count") val questionCount: Int,
    val status: String = "in_progress"
)

@Serializable
data class PracticeQuestionDto(
    val id: String,
    @SerialName("question_type") val questionType: String = "mcq_single",
    val difficulty: String = "medium",
    val body: String,
    val options: List<OptionForPlayer> = emptyList(),
    @SerialName("order_index") val orderIndex: Int = 1
)

@Serializable
data class SubmitPracticeAnswerRequestDto(
    @SerialName("question_id") val questionId: String,
    @SerialName("selected_option_id") val selectedOptionId: String,
    @SerialName("time_taken_ms") val timeTakenMs: Int = 0
)

@Serializable
data class SubmitPracticeAnswerResponseDto(
    @SerialName("is_correct") val isCorrect: Boolean,
    @SerialName("correct_option_id") val correctOptionId: String? = null,
    val explanation: String? = null
)
