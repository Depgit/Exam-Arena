package com.examarena.feature.splash

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import com.examarena.core.common.Resource
import com.examarena.data.repository.AuthRepository
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch

sealed interface SplashDestination {
    data object Loading : SplashDestination
    data object Home : SplashDestination
    data object Onboarding : SplashDestination
    data object Auth : SplashDestination
}

class SplashViewModel(
    private val authRepository: AuthRepository
) : ViewModel() {

    private val _destination = MutableStateFlow<SplashDestination>(SplashDestination.Loading)
    val destination: StateFlow<SplashDestination> = _destination.asStateFlow()

    init {
        checkSession()
    }

    private fun checkSession() {
        viewModelScope.launch {
            // Guarantee at least 1.2 seconds of splash display for smooth brand animation
            delay(1200)

            val token = authRepository.authTokenFlow.first()
            if (!token.isNullOrBlank()) {
                val userResult = authRepository.getCurrentUser()
                if (userResult is Resource.Success) {
                    _destination.value = SplashDestination.Home
                } else {
                    _destination.value = SplashDestination.Auth
                }
            } else {
                _destination.value = SplashDestination.Onboarding
            }
        }
    }

    companion object {
        fun provideFactory(authRepository: AuthRepository): ViewModelProvider.Factory =
            object : ViewModelProvider.Factory {
                @Suppress("UNCHECKED_CAST")
                override fun <T : ViewModel> create(modelClass: Class<T>): T {
                    return SplashViewModel(authRepository) as T
                }
            }
    }
}
