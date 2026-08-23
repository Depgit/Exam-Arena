package com.examarena.feature.practice

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import com.examarena.core.common.Resource
import com.examarena.data.model.ExamCategory
import com.examarena.data.model.PracticeQuestionDto
import com.examarena.data.model.PracticeSessionDto
import com.examarena.data.model.SubmitPracticeAnswerResponseDto
import com.examarena.data.repository.PracticeRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class PracticeUiState(
    val selectedCategory: ExamCategory = ExamCategory.SSCCGL,
    val selectedQuestionCount: Int = 5,
    val session: PracticeSessionDto? = null,
    val questions: List<PracticeQuestionDto> = emptyList(),
    val currentQuestionIndex: Int = 0,
    val selectedOptionId: String? = null,
    val evaluationResult: SubmitPracticeAnswerResponseDto? = null,
    val isAnswerEvaluated: Boolean = false,
    val correctAnswersCount: Int = 0,
    val isSessionCompleted: Boolean = false,
    val isLoading: Boolean = false,
    val errorMessage: String? = null
)

class PracticeViewModel(
    private val practiceRepository: PracticeRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(PracticeUiState())
    val uiState: StateFlow<PracticeUiState> = _uiState.asStateFlow()

    private var questionStartTimeMs: Long = System.currentTimeMillis()

    fun selectCategory(category: ExamCategory) {
        _uiState.value = _uiState.value.copy(selectedCategory = category)
    }

    fun selectQuestionCount(count: Int) {
        _uiState.value = _uiState.value.copy(selectedQuestionCount = count)
    }

    fun startPracticeSession(onSessionStarted: () -> Unit) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, errorMessage = null)
            val state = _uiState.value
            val result = practiceRepository.startPractice(
                examCategoryId = state.selectedCategory.id,
                difficulty = "mixed",
                questionCount = state.selectedQuestionCount
            )

            if (result is Resource.Success) {
                _uiState.value = _uiState.value.copy(
                    session = result.data.session,
                    questions = result.data.questions,
                    currentQuestionIndex = 0,
                    selectedOptionId = null,
                    evaluationResult = null,
                    isAnswerEvaluated = false,
                    correctAnswersCount = 0,
                    isSessionCompleted = false,
                    isLoading = false
                )
                questionStartTimeMs = System.currentTimeMillis()
                onSessionStarted()
            } else if (result is Resource.Error) {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    errorMessage = result.message
                )
            }
        }
    }

    fun submitAnswer(optionId: String) {
        val state = _uiState.value
        val session = state.session ?: return
        val currentQuestion = state.questions.getOrNull(state.currentQuestionIndex) ?: return

        if (state.isAnswerEvaluated) return

        val timeTakenMs = (System.currentTimeMillis() - questionStartTimeMs).toInt().coerceAtLeast(100)

        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(selectedOptionId = optionId, isLoading = true)
            val result = practiceRepository.submitAnswer(
                sessionId = session.id,
                questionId = currentQuestion.id,
                selectedOptionId = optionId,
                timeTakenMs = timeTakenMs
            )

            if (result is Resource.Success) {
                val eval = result.data
                val newCorrectCount = if (eval.isCorrect) state.correctAnswersCount + 1 else state.correctAnswersCount
                _uiState.value = _uiState.value.copy(
                    evaluationResult = eval,
                    isAnswerEvaluated = true,
                    correctAnswersCount = newCorrectCount,
                    isLoading = false
                )
            } else if (result is Resource.Error) {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    errorMessage = result.message
                )
            }
        }
    }

    fun nextQuestion() {
        val state = _uiState.value
        if (state.currentQuestionIndex < state.questions.size - 1) {
            _uiState.value = state.copy(
                currentQuestionIndex = state.currentQuestionIndex + 1,
                selectedOptionId = null,
                evaluationResult = null,
                isAnswerEvaluated = false
            )
            questionStartTimeMs = System.currentTimeMillis()
        } else {
            // End practice session
            val session = state.session
            if (session != null) {
                viewModelScope.launch {
                    practiceRepository.endSession(session.id)
                }
            }
            _uiState.value = state.copy(isSessionCompleted = true)
        }
    }

    companion object {
        fun provideFactory(practiceRepository: PracticeRepository): ViewModelProvider.Factory =
            object : ViewModelProvider.Factory {
                @Suppress("UNCHECKED_CAST")
                override fun <T : ViewModel> create(modelClass: Class<T>): T {
                    return PracticeViewModel(practiceRepository) as T
                }
            }
    }
}
