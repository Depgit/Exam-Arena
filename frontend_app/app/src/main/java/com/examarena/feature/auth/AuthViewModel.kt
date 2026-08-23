package com.examarena.feature.auth

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import com.examarena.core.common.Resource
import com.examarena.data.repository.AuthRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import kotlin.random.Random

data class AuthUiState(
    val isLoginTab: Boolean = true,
    val loginIdentifier: String = "",
    val registerUsername: String = "",
    val registerEmail: String = "",
    val password: String = "",
    val confirmPassword: String = "",
    val isLoading: Boolean = false,
    val errorMessage: String? = null,
    val isSuccess: Boolean = false
)

class AuthViewModel(
    private val authRepository: AuthRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(AuthUiState())
    val uiState: StateFlow<AuthUiState> = _uiState.asStateFlow()

    fun switchTab(isLogin: Boolean) {
        _uiState.value = _uiState.value.copy(
            isLoginTab = isLogin,
            errorMessage = null
        )
    }

    fun onLoginIdentifierChange(value: String) {
        _uiState.value = _uiState.value.copy(loginIdentifier = value, errorMessage = null)
    }

    fun onRegisterUsernameChange(value: String) {
        _uiState.value = _uiState.value.copy(registerUsername = value, errorMessage = null)
    }

    fun onRegisterEmailChange(value: String) {
        _uiState.value = _uiState.value.copy(registerEmail = value, errorMessage = null)
    }

    fun onPasswordChange(value: String) {
        _uiState.value = _uiState.value.copy(password = value, errorMessage = null)
    }

    fun onConfirmPasswordChange(value: String) {
        _uiState.value = _uiState.value.copy(confirmPassword = value, errorMessage = null)
    }

    fun submitLogin() {
        val state = _uiState.value
        if (state.loginIdentifier.isBlank() || state.password.isBlank()) {
            _uiState.value = state.copy(errorMessage = "Please enter both login and password.")
            return
        }

        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, errorMessage = null)
            val result = authRepository.login(state.loginIdentifier.trim(), state.password)
            when (result) {
                is Resource.Success -> {
                    _uiState.value = _uiState.value.copy(isLoading = false, isSuccess = true)
                }
                is Resource.Error -> {
                    _uiState.value = _uiState.value.copy(isLoading = false, errorMessage = result.message)
                }
                else -> Unit
            }
        }
    }

    fun submitRegister() {
        val state = _uiState.value
        if (state.registerUsername.length < 3) {
            _uiState.value = state.copy(errorMessage = "Username must be at least 3 characters.")
            return
        }
        if (!state.registerEmail.contains("@") || !state.registerEmail.contains(".")) {
            _uiState.value = state.copy(errorMessage = "Please enter a valid email address.")
            return
        }
        if (state.password.length < 8) {
            _uiState.value = state.copy(errorMessage = "Password must be at least 8 characters.")
            return
        }
        if (state.password != state.confirmPassword) {
            _uiState.value = state.copy(errorMessage = "Passwords do not match.")
            return
        }

        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, errorMessage = null)
            val result = authRepository.register(
                username = state.registerUsername.trim(),
                email = state.registerEmail.trim().lowercase(),
                password = state.password
            )
            when (result) {
                is Resource.Success -> {
                    _uiState.value = _uiState.value.copy(isLoading = false, isSuccess = true)
                }
                is Resource.Error -> {
                    _uiState.value = _uiState.value.copy(isLoading = false, errorMessage = result.message)
                }
                else -> Unit
            }
        }
    }

    fun quickGuestLogin() {
        val randomSuffix = Random.nextInt(1000, 9999)
        val guestUsername = "player_$randomSuffix"
        val guestEmail = "player$randomSuffix@examarena.local"
        val guestPassword = "Password123!"

        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, errorMessage = null)
            val result = authRepository.register(guestUsername, guestEmail, guestPassword)
            when (result) {
                is Resource.Success -> {
                    _uiState.value = _uiState.value.copy(isLoading = false, isSuccess = true)
                }
                is Resource.Error -> {
                    // Try login if already registered
                    val loginResult = authRepository.login(guestUsername, guestPassword)
                    if (loginResult is Resource.Success) {
                        _uiState.value = _uiState.value.copy(isLoading = false, isSuccess = true)
                    } else {
                        _uiState.value = _uiState.value.copy(isLoading = false, errorMessage = result.message)
                    }
                }
                else -> Unit
            }
        }
    }

    companion object {
        fun provideFactory(authRepository: AuthRepository): ViewModelProvider.Factory =
            object : ViewModelProvider.Factory {
                @Suppress("UNCHECKED_CAST")
                override fun <T : ViewModel> create(modelClass: Class<T>): T {
                    return AuthViewModel(authRepository) as T
                }
            }
    }
}
