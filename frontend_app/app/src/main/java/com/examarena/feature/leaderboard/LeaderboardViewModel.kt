package com.examarena.feature.leaderboard

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import com.examarena.core.common.Resource
import com.examarena.data.model.ExamCategory
import com.examarena.data.model.LeaderboardEntryDto
import com.examarena.data.repository.LeaderboardRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class LeaderboardUiState(
    val selectedCategory: ExamCategory = ExamCategory.SSCCGL,
    val entries: List<LeaderboardEntryDto> = emptyList(),
    val isLoading: Boolean = false,
    val errorMessage: String? = null
)

class LeaderboardViewModel(
    private val leaderboardRepository: LeaderboardRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(LeaderboardUiState())
    val uiState: StateFlow<LeaderboardUiState> = _uiState.asStateFlow()

    init {
        loadLeaderboard(ExamCategory.SSCCGL)
    }

    fun selectCategory(category: ExamCategory) {
        _uiState.value = _uiState.value.copy(selectedCategory = category)
        loadLeaderboard(category)
    }

    fun loadLeaderboard(category: ExamCategory = _uiState.value.selectedCategory) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, errorMessage = null)
            val result = leaderboardRepository.getLeaderboard(category.id)
            if (result is Resource.Success) {
                _uiState.value = _uiState.value.copy(
                    entries = result.data,
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
        fun provideFactory(leaderboardRepository: LeaderboardRepository): ViewModelProvider.Factory =
            object : ViewModelProvider.Factory {
                @Suppress("UNCHECKED_CAST")
                override fun <T : ViewModel> create(modelClass: Class<T>): T {
                    return LeaderboardViewModel(leaderboardRepository) as T
                }
            }
    }
}
