package com.examarena.feature.matchmaking

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import com.examarena.core.common.Resource
import com.examarena.core.websocket.WsEvent
import com.examarena.data.model.ExamCategory
import com.examarena.data.repository.MatchRepository
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch

data class MatchmakingUiState(
    val category: ExamCategory = ExamCategory.SSCCGL,
    val matchType: String = "ranked",
    val elapsedSeconds: Int = 0,
    val ratingRange: String = "1200 – 1300",
    val isQueued: Boolean = false,
    val matchedMatchId: String? = null,
    val errorMessage: String? = null
)

class MatchmakingViewModel(
    private val categoryId: String,
    private val matchType: String,
    private val matchRepository: MatchRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(
        MatchmakingUiState(
            category = ExamCategory.fromId(categoryId),
            matchType = matchType
        )
    )
    val uiState: StateFlow<MatchmakingUiState> = _uiState.asStateFlow()

    private var timerJob: Job? = null

    init {
        startMatchmaking()
        observeWebSocketEvents()
    }

    private fun startMatchmaking() {
        viewModelScope.launch {
            // Ensure WebSocket connection is ready
            matchRepository.connectWebSocket()

            // Join backend matchmaking queue
            val result = matchRepository.joinQueue(categoryId, matchType)
            if (result is Resource.Success) {
                _uiState.value = _uiState.value.copy(isQueued = true, errorMessage = null)
                startSearchTimer()
            } else if (result is Resource.Error) {
                _uiState.value = _uiState.value.copy(errorMessage = result.message)
            }
        }
    }

    private fun startSearchTimer() {
        timerJob?.cancel()
        timerJob = viewModelScope.launch {
            var seconds = 0
            while (isActive) {
                delay(1000)
                seconds++
                val range = when {
                    seconds < 6 -> "±100 Rating"
                    seconds < 12 -> "±150 Rating"
                    seconds < 20 -> "±250 Rating"
                    else -> "±400 Rating (Any Opponent)"
                }
                _uiState.value = _uiState.value.copy(
                    elapsedSeconds = seconds,
                    ratingRange = range
                )
            }
        }
    }

    private fun observeWebSocketEvents() {
        viewModelScope.launch {
            matchRepository.wsEvents.collect { event ->
                when (event) {
                    is WsEvent.MatchStart -> {
                        timerJob?.cancel()
                        _uiState.value = _uiState.value.copy(matchedMatchId = event.matchId)
                    }
                    is WsEvent.MatchFailed -> {
                        timerJob?.cancel()
                        _uiState.value = _uiState.value.copy(
                            isQueued = false,
                            errorMessage = event.reason
                        )
                    }
                    is WsEvent.Error -> {
                        _uiState.value = _uiState.value.copy(errorMessage = event.message)
                    }
                    else -> Unit
                }
            }
        }
    }

    fun cancelMatchmaking(onSuccess: () -> Unit) {
        viewModelScope.launch {
            timerJob?.cancel()
            matchRepository.leaveQueue()
            onSuccess()
        }
    }

    override fun onCleared() {
        super.onCleared()
        timerJob?.cancel()
    }

    companion object {
        fun provideFactory(
            categoryId: String,
            matchType: String,
            matchRepository: MatchRepository
        ): ViewModelProvider.Factory = object : ViewModelProvider.Factory {
            @Suppress("UNCHECKED_CAST")
            override fun <T : ViewModel> create(modelClass: Class<T>): T {
                return MatchmakingViewModel(categoryId, matchType, matchRepository) as T
            }
        }
    }
}
