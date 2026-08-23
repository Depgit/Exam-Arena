package com.examarena.feature.profile

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import com.examarena.core.common.Resource
import com.examarena.core.datastore.SessionManager
import com.examarena.data.model.UserDto
import com.examarena.data.model.UserRatingDto
import com.examarena.data.model.UserStatisticsDto
import com.examarena.data.repository.AuthRepository
import com.examarena.data.repository.UserRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch

data class ProfileUiState(
    val user: UserDto? = null,
    val ratings: List<UserRatingDto> = emptyList(),
    val statistics: UserStatisticsDto? = null,
    val isLoading: Boolean = false,
    val errorMessage: String? = null
)

class ProfileViewModel(
    private val authRepository: AuthRepository,
    private val userRepository: UserRepository,
    private val sessionManager: SessionManager
) : ViewModel() {

    private val _uiState = MutableStateFlow(ProfileUiState())
    val uiState: StateFlow<ProfileUiState> = _uiState.asStateFlow()

    val apiUrlFlow: StateFlow<String> = sessionManager.apiUrlFlow.stateIn(
        viewModelScope,
        SharingStarted.WhileSubscribed(5000),
        SessionManager.DEFAULT_HTTP_URL
    )

    val wsUrlFlow: StateFlow<String> = sessionManager.wsUrlFlow.stateIn(
        viewModelScope,
        SharingStarted.WhileSubscribed(5000),
        SessionManager.DEFAULT_WS_URL
    )

    val isDarkThemeFlow: StateFlow<Boolean?> = sessionManager.isDarkThemeFlow.stateIn(
        viewModelScope,
        SharingStarted.WhileSubscribed(5000),
        null
    )

    val isSoundEnabledFlow: StateFlow<Boolean> = sessionManager.soundEnabledFlow.stateIn(
        viewModelScope,
        SharingStarted.WhileSubscribed(5000),
        true
    )

    init {
        loadProfileData()
    }

    fun loadProfileData() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, errorMessage = null)
            val userResult = authRepository.getCurrentUser()

            if (userResult is Resource.Success) {
                val user = userResult.data
                _uiState.value = _uiState.value.copy(user = user)

                val profileResult = userRepository.getUserProfile(user.id)
                if (profileResult is Resource.Success) {
                    _uiState.value = _uiState.value.copy(
                        ratings = profileResult.data.ratings
                    )
                }

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

    fun updateServerConfig(apiUrl: String, wsUrl: String) {
        viewModelScope.launch {
            sessionManager.updateServerUrls(apiUrl.trim(), wsUrl.trim())
        }
    }

    fun toggleDarkMode(enabled: Boolean?) {
        viewModelScope.launch {
            sessionManager.setDarkTheme(enabled)
        }
    }

    fun toggleSound(enabled: Boolean) {
        viewModelScope.launch {
            sessionManager.setSoundEnabled(enabled)
        }
    }

    fun logout(onLoggedOut: () -> Unit) {
        viewModelScope.launch {
            authRepository.logout()
            onLoggedOut()
        }
    }

    companion object {
        fun provideFactory(
            authRepository: AuthRepository,
            userRepository: UserRepository,
            sessionManager: SessionManager
        ): ViewModelProvider.Factory = object : ViewModelProvider.Factory {
            @Suppress("UNCHECKED_CAST")
            override fun <T : ViewModel> create(modelClass: Class<T>): T {
                return ProfileViewModel(authRepository, userRepository, sessionManager) as T
            }
        }
    }
}
