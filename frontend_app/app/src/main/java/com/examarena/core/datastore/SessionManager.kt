package com.examarena.core.datastore

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.booleanPreferencesKey
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map

private val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = "exam_arena_prefs")

class SessionManager(private val context: Context) {

    companion object {
        private val KEY_AUTH_TOKEN = stringPreferencesKey("auth_token")
        private val KEY_USER_ID = stringPreferencesKey("user_id")
        private val KEY_USERNAME = stringPreferencesKey("username")
        private val KEY_EMAIL = stringPreferencesKey("email")
        private val KEY_DISPLAY_NAME = stringPreferencesKey("display_name")
        private val KEY_ROLE = stringPreferencesKey("role")
        private val KEY_API_URL = stringPreferencesKey("api_url")
        private val KEY_WS_URL = stringPreferencesKey("ws_url")
        private val KEY_DARK_THEME = booleanPreferencesKey("dark_theme")
        private val KEY_SOUND_ENABLED = booleanPreferencesKey("sound_enabled")
        private val KEY_HAPTICS_ENABLED = booleanPreferencesKey("haptics_enabled")

        // Default URLs: 10.0.2.2 is Android emulator's alias for host loopback interface
        const val DEFAULT_HTTP_URL = "http://10.0.2.2:8080"
        const val DEFAULT_WS_URL = "ws://10.0.2.2:8080/ws"
    }

    val authTokenFlow: Flow<String?> = context.dataStore.data.map { it[KEY_AUTH_TOKEN] }
    val userIdFlow: Flow<String?> = context.dataStore.data.map { it[KEY_USER_ID] }
    val usernameFlow: Flow<String?> = context.dataStore.data.map { it[KEY_USERNAME] }
    val emailFlow: Flow<String?> = context.dataStore.data.map { it[KEY_EMAIL] }
    val apiUrlFlow: Flow<String> = context.dataStore.data.map { it[KEY_API_URL] ?: DEFAULT_HTTP_URL }
    val wsUrlFlow: Flow<String> = context.dataStore.data.map { it[KEY_WS_URL] ?: DEFAULT_WS_URL }
    val isDarkThemeFlow: Flow<Boolean?> = context.dataStore.data.map { it[KEY_DARK_THEME] }
    val soundEnabledFlow: Flow<Boolean> = context.dataStore.data.map { it[KEY_SOUND_ENABLED] ?: true }
    val hapticsEnabledFlow: Flow<Boolean> = context.dataStore.data.map { it[KEY_HAPTICS_ENABLED] ?: true }

    suspend fun saveAuthSession(
        token: String,
        userId: String,
        username: String,
        email: String,
        displayName: String? = null,
        role: String = "user"
    ) {
        context.dataStore.edit { prefs ->
            prefs[KEY_AUTH_TOKEN] = token
            prefs[KEY_USER_ID] = userId
            prefs[KEY_USERNAME] = username
            prefs[KEY_EMAIL] = email
            displayName?.let { prefs[KEY_DISPLAY_NAME] = it }
            prefs[KEY_ROLE] = role
        }
    }

    suspend fun getAuthToken(): String? {
        return context.dataStore.data.map { it[KEY_AUTH_TOKEN] }.first()
    }

    suspend fun getUserId(): String? {
        return context.dataStore.data.map { it[KEY_USER_ID] }.first()
    }

    suspend fun getUsername(): String? {
        return context.dataStore.data.map { it[KEY_USERNAME] }.first()
    }

    suspend fun getApiUrl(): String {
        return context.dataStore.data.map { it[KEY_API_URL] ?: DEFAULT_HTTP_URL }.first()
    }

    suspend fun getWsUrl(): String {
        return context.dataStore.data.map { it[KEY_WS_URL] ?: DEFAULT_WS_URL }.first()
    }

    suspend fun updateServerUrls(apiUrl: String, wsUrl: String) {
        context.dataStore.edit { prefs ->
            prefs[KEY_API_URL] = apiUrl
            prefs[KEY_WS_URL] = wsUrl
        }
    }

    suspend fun setDarkTheme(enabled: Boolean?) {
        context.dataStore.edit { prefs ->
            if (enabled == null) {
                prefs.remove(KEY_DARK_THEME)
            } else {
                prefs[KEY_DARK_THEME] = enabled
            }
        }
    }

    suspend fun setSoundEnabled(enabled: Boolean) {
        context.dataStore.edit { it[KEY_SOUND_ENABLED] = enabled }
    }

    suspend fun setHapticsEnabled(enabled: Boolean) {
        context.dataStore.edit { it[KEY_HAPTICS_ENABLED] = enabled }
    }

    suspend fun clearSession() {
        context.dataStore.edit { prefs ->
            prefs.remove(KEY_AUTH_TOKEN)
            prefs.remove(KEY_USER_ID)
            prefs.remove(KEY_USERNAME)
            prefs.remove(KEY_EMAIL)
            prefs.remove(KEY_DISPLAY_NAME)
            prefs.remove(KEY_ROLE)
        }
    }
}
