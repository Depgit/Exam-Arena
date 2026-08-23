package com.examarena.feature.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import com.examarena.core.common.Resource
import com.examarena.data.model.ExamCategory
import com.examarena.data.model.UserDto
import com.examarena.data.model.UserRatingDto
import com.examarena.data.model.UserStatisticsDto
import com.examarena.data.repository.AuthRepository
import com.examarena.data.repository.UserRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch

data class HomeUiState(
    val user: UserDto? = null,
    val selectedCategory: ExamCategory = ExamCategory.SSCCGL,
    val currentRating: Int = 1200,
    val ratingRank: String = "#1",
    val statistics: UserStatisticsDto? = null,
    val ratingsList: List<UserRatingDto> = emptyList(),
    val isLoading: Boolean = false,
    val errorMessage: String? = null
)

class HomeViewModel(
    private val authRepository: AuthRepository,
    private val userRepository: UserRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(HomeUiState())
    val uiState: StateFlow<HomeUiState> = _uiState.asStateFlow()

    init {
        loadDashboardData()
    }

    fun selectCategory(category: ExamCategory) {
        val currentRatings = _uiState.value.ratingsList
        val categoryRating = currentRatings.find { it.examCategoryId == category.id }?.rating ?: 1200
        _uiState.value = _uiState.value.copy(
            selectedCategory = category,
            currentRating = categoryRating
        )
    }

    fun loadDashboardData() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, errorMessage = null)
            val userResult = authRepository.getCurrentUser()

            if (userResult is Resource.Success) {
                val user = userResult.data
                _uiState.value = _uiState.value.copy(user = user)

                // Fetch public profile with ratings
                val profileResult = userRepository.getUserProfile(user.id)
                if (profileResult is Resource.Success) {
                    val ratings = profileResult.data.ratings
                    val currentCategoryRating = ratings.find {
                        it.examCategoryId == _uiState.value.selectedCategory.id
                    }?.rating ?: 1200

                    _uiState.value = _uiState.value.copy(
                        ratingsList = ratings,
                        currentRating = currentCategoryRating
                    )
                }

                // Fetch user lifetime stats
                val statsResult = userRepository.getUserStatistics(user.id)
                if (statsResult is Resource.Success) {
                    _uiState.value = _uiState.value.copy(
                        statistics = statsResult.data
                    )
                }

                _uiState.value = _uiState.value.copy(isLoading = false)
            } else if (userResult is Resource.Error) {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    errorMessage = userResult.message
                )
            }
        }
    }

    companion object {
        fun provideFactory(
            authRepository: AuthRepository,
            userRepository: UserRepository
        ): ViewModelProvider.Factory = object : ViewModelProvider.Factory {
            @Suppress("UNCHECKED_CAST")
            override fun <T : ViewModel> create(modelClass: Class<T>): T {
                return HomeViewModel(authRepository, userRepository) as T
            }
        }
    }
}
