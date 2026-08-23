package com.examarena.data.repository

import com.examarena.core.common.Resource
import com.examarena.core.datastore.SessionManager
import com.examarena.core.network.ApiClient
import com.examarena.data.model.AuthResponseDto
import com.examarena.data.model.FirebaseLoginRequestDto
import com.examarena.data.model.LoginRequestDto
import com.examarena.data.model.RegisterRequestDto
import com.examarena.data.model.UserDto
import kotlinx.coroutines.flow.Flow
import kotlinx.serialization.encodeToString

class AuthRepository(
    private val apiClient: ApiClient,
    private val sessionManager: SessionManager
) {
    val authTokenFlow: Flow<String?> = sessionManager.authTokenFlow
    val userIdFlow: Flow<String?> = sessionManager.userIdFlow
    val usernameFlow: Flow<String?> = sessionManager.usernameFlow

    suspend fun register(username: String, email: String, password: String): Resource<AuthResponseDto> {
        val requestDto = RegisterRequestDto(username, email, password)
        val bodyJson = apiClient.json.encodeToString(requestDto)

        val result = apiClient.post("/api/v1/auth/register", bodyJson, requireAuth = false) { body ->
            apiClient.json.decodeFromString<AuthResponseDto>(body)
        }

        if (result is Resource.Success) {
            sessionManager.saveAuthSession(
                token = result.data.token,
                userId = result.data.user.id,
                username = result.data.user.username,
                email = result.data.user.email ?: email,
                displayName = result.data.user.displayName,
                role = result.data.user.role
            )
        }

        return result
    }

    suspend fun login(login: String, password: String): Resource<AuthResponseDto> {
        val requestDto = LoginRequestDto(login, password)
        val bodyJson = apiClient.json.encodeToString(requestDto)

        val result = apiClient.post("/api/v1/auth/login", bodyJson, requireAuth = false) { body ->
            apiClient.json.decodeFromString<AuthResponseDto>(body)
        }

        if (result is Resource.Success) {
            sessionManager.saveAuthSession(
                token = result.data.token,
                userId = result.data.user.id,
                username = result.data.user.username,
                email = result.data.user.email ?: "",
                displayName = result.data.user.displayName,
                role = result.data.user.role
            )
        }

        return result
    }

    suspend fun loginWithFirebase(idToken: String, email: String, username: String? = null): Resource<AuthResponseDto> {
        val requestDto = FirebaseLoginRequestDto(idToken, email, username)
        val bodyJson = apiClient.json.encodeToString(requestDto)

        val result = apiClient.post("/api/v1/auth/firebase", bodyJson, requireAuth = false) { body ->
            apiClient.json.decodeFromString<AuthResponseDto>(body)
        }

        if (result is Resource.Success) {
            sessionManager.saveAuthSession(
                token = result.data.token,
                userId = result.data.user.id,
                username = result.data.user.username,
                email = result.data.user.email ?: email,
                displayName = result.data.user.displayName,
                role = result.data.user.role
            )
        }

        return result
    }

    suspend fun getCurrentUser(): Resource<UserDto> {
        return apiClient.get("/api/v1/auth/me", requireAuth = true) { body ->
            apiClient.json.decodeFromString<UserDto>(body)
        }
    }

    suspend fun logout() {
        sessionManager.clearSession()
    }
}
