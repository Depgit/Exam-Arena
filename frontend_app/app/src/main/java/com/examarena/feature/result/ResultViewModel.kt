package com.examarena.feature.result

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import com.examarena.core.common.Resource
import com.examarena.data.model.MatchDetailsResponseDto
import com.examarena.data.model.MatchPlayerDto
import com.examarena.data.repository.MatchRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class ResultUiState(
    val matchId: String = "",
    val isWinner: Boolean = false,
    val isDraw: Boolean = false,
    val myPlayer: MatchPlayerDto? = null,
    val opponentPlayer: MatchPlayerDto? = null,
    val ratingDelta: Int = 32,
    val correctCount: Int = 0,
    val totalQuestions: Int = 10,
    val accuracyPercent: Int = 0,
    val isLoading: Boolean = true,
    val errorMessage: String? = null
)

class ResultViewModel(
    private val matchId: String,
    private val currentUserId: String?,
    private val matchRepository: MatchRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(ResultUiState(matchId = matchId))
    val uiState: StateFlow<ResultUiState> = _uiState.asStateFlow()

    init {
        loadResultDetails()
    }

    private fun loadResultDetails() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true)
            val result = matchRepository.getMatchDetails(matchId)

            if (result is Resource.Success) {
                val details = result.data
                val myPlayer = details.players.find { it.userId == currentUserId } ?: details.players.firstOrNull()
                val opponentPlayer = details.players.find { it.userId != currentUserId } ?: details.players.lastOrNull()

                val myScore = myPlayer?.score ?: 0
                val opponentScore = opponentPlayer?.score ?: 0
                val isWinner = myScore > opponentScore
                val isDraw = myScore == opponentScore

                val delta = myPlayer?.ratingDelta ?: if (isWinner) 32 else if (isDraw) 0 else -18
                val totalQ = details.questions.size.coerceAtLeast(1)
                val correctEstimate = (myScore / 100).coerceAtMost(totalQ)
                val accuracy = ((correctEstimate.toFloat() / totalQ.toFloat()) * 100).toInt()

                _uiState.value = _uiState.value.copy(
                    isWinner = isWinner,
                    isDraw = isDraw,
                    myPlayer = myPlayer,
                    opponentPlayer = opponentPlayer,
                    ratingDelta = delta,
                    correctCount = correctEstimate,
                    totalQuestions = totalQ,
                    accuracyPercent = accuracy,
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

    companion object {
        fun provideFactory(
            matchId: String,
            currentUserId: String?,
            matchRepository: MatchRepository
        ): ViewModelProvider.Factory = object : ViewModelProvider.Factory {
            @Suppress("UNCHECKED_CAST")
            override fun <T : ViewModel> create(modelClass: Class<T>): T {
                return ResultViewModel(matchId, currentUserId, matchRepository) as T
            }
        }
    }
}
