package com.examarena.feature.battle

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import com.examarena.core.common.Resource
import com.examarena.core.websocket.WsEvent
import com.examarena.data.model.OptionForPlayer
import com.examarena.data.model.QuestionForPlayer
import com.examarena.data.model.ScoreboardEntry
import com.examarena.data.repository.MatchRepository
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch

data class BattleUiState(
    val matchId: String = "",
    val questions: List<QuestionForPlayer> = emptyList(),
    val currentQuestionIndex: Int = 0,
    val selectedOptionId: String? = null,
    val isAnswerSubmitted: Boolean = false,
    val myScore: Int = 0,
    val opponentScore: Int = 0,
    val opponentUsername: String = "Opponent",
    val isOpponentAnsweredCurrent: Boolean = false,
    val remainingSeconds: Int = 120,
    val totalTimerSeconds: Int = 120,
    val isMatchCompleted: Boolean = false,
    val isLoading: Boolean = true,
    val errorMessage: String? = null
)

class BattleViewModel(
    private val matchId: String,
    private val currentUserId: String?,
    private val matchRepository: MatchRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(BattleUiState(matchId = matchId))
    val uiState: StateFlow<BattleUiState> = _uiState.asStateFlow()

    private var questionStartTimeMs: Long = System.currentTimeMillis()
    private var localTimerJob: Job? = null

    init {
        loadMatchDetails()
        observeWebSocketEvents()
        startLocalTimer()
    }

    private fun loadMatchDetails() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true)
            val result = matchRepository.getMatchDetails(matchId)
            if (result is Resource.Success) {
                val details = result.data
                val opponent = details.players.find { it.userId != currentUserId }
                _uiState.value = _uiState.value.copy(
                    questions = details.questions,
                    totalTimerSeconds = details.match.timerSeconds,
                    remainingSeconds = details.match.timerSeconds,
                    opponentUsername = opponent?.username ?: "Opponent",
                    isLoading = false
                )
                questionStartTimeMs = System.currentTimeMillis()
            } else if (result is Resource.Error) {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    errorMessage = result.message
                )
            }
        }
    }

    private fun observeWebSocketEvents() {
        viewModelScope.launch {
            matchRepository.wsEvents.collect { event ->
                when (event) {
                    is WsEvent.MatchStart -> {
                        if (event.matchId == matchId) {
                            val opponent = event.players.find { it.userId != currentUserId }
                            _uiState.value = _uiState.value.copy(
                                questions = event.questions,
                                totalTimerSeconds = event.timerSeconds,
                                remainingSeconds = event.timerSeconds,
                                opponentUsername = opponent?.username ?: "Opponent",
                                isLoading = false
                            )
                            questionStartTimeMs = System.currentTimeMillis()
                        }
                    }
                    is WsEvent.ScoreUpdate -> {
                        if (event.matchId == matchId) {
                            updateScoreboard(event.scoreboard, event.userId)
                        }
                    }
                    is WsEvent.TimeUpdate -> {
                        if (event.matchId == matchId) {
                            _uiState.value = _uiState.value.copy(
                                remainingSeconds = event.remainingSeconds
                            )
                        }
                    }
                    is WsEvent.MatchEnd -> {
                        if (event.matchId == matchId) {
                            localTimerJob?.cancel()
                            _uiState.value = _uiState.value.copy(isMatchCompleted = true)
                        }
                    }
                    is WsEvent.Error -> {
                        _uiState.value = _uiState.value.copy(errorMessage = event.message)
                    }
                    else -> Unit
                }
            }
        }
    }

    private fun startLocalTimer() {
        localTimerJob?.cancel()
        localTimerJob = viewModelScope.launch {
            while (isActive) {
                delay(1000)
                val current = _uiState.value.remainingSeconds
                if (current > 0) {
                    _uiState.value = _uiState.value.copy(remainingSeconds = current - 1)
                } else {
                    break
                }
            }
        }
    }

    private fun updateScoreboard(scoreboard: List<ScoreboardEntry>, lastAnswerUserId: String) {
        val myEntry = scoreboard.find { it.userId == currentUserId }
        val opponentEntry = scoreboard.find { it.userId != currentUserId }

        val isOpponent = lastAnswerUserId != currentUserId

        _uiState.value = _uiState.value.copy(
            myScore = myEntry?.score ?: _uiState.value.myScore,
            opponentScore = opponentEntry?.score ?: _uiState.value.opponentScore,
            isOpponentAnsweredCurrent = isOpponent || _uiState.value.isOpponentAnsweredCurrent
        )
    }

    fun selectOption(option: OptionForPlayer) {
        val state = _uiState.value
        if (state.isAnswerSubmitted || state.currentQuestionIndex >= state.questions.size) return

        val currentQuestion = state.questions[state.currentQuestionIndex]
        val timeTakenMs = (System.currentTimeMillis() - questionStartTimeMs).toInt().coerceAtLeast(100)

        _uiState.value = state.copy(
            selectedOptionId = option.id,
            isAnswerSubmitted = true
        )

        // Submit via WebSocket
        matchRepository.submitAnswer(
            matchId = matchId,
            questionId = currentQuestion.id,
            optionId = option.id,
            timeTakenMs = timeTakenMs
        )

        // Advance to next question after short tactical delay (400ms)
        viewModelScope.launch {
            delay(400)
            nextQuestion()
        }
    }

    private fun nextQuestion() {
        val state = _uiState.value
        if (state.currentQuestionIndex < state.questions.size - 1) {
            _uiState.value = state.copy(
                currentQuestionIndex = state.currentQuestionIndex + 1,
                selectedOptionId = null,
                isAnswerSubmitted = false,
                isOpponentAnsweredCurrent = false
            )
            questionStartTimeMs = System.currentTimeMillis()
        } else {
            // All questions finished, wait for server match_end event
            _uiState.value = state.copy(isAnswerSubmitted = true)
        }
    }

    override fun onCleared() {
        super.onCleared()
        localTimerJob?.cancel()
    }

    companion object {
        fun provideFactory(
            matchId: String,
            currentUserId: String?,
            matchRepository: MatchRepository
        ): ViewModelProvider.Factory = object : ViewModelProvider.Factory {
            @Suppress("UNCHECKED_CAST")
            override fun <T : ViewModel> create(modelClass: Class<T>): T {
                return BattleViewModel(matchId, currentUserId, matchRepository) as T
            }
        }
    }
}
