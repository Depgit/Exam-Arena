package com.examarena.core.di

import android.content.Context
import com.examarena.core.datastore.SessionManager
import com.examarena.core.network.ApiClient
import com.examarena.core.websocket.ExamArenaWebSocket
import com.examarena.data.repository.AuthRepository
import com.examarena.data.repository.LeaderboardRepository
import com.examarena.data.repository.MatchRepository
import com.examarena.data.repository.PracticeRepository
import com.examarena.data.repository.UserRepository

class AppContainer(context: Context) {
    val sessionManager: SessionManager by lazy {
        SessionManager(context.applicationContext)
    }

    val apiClient: ApiClient by lazy {
        ApiClient(sessionManager)
    }

    val webSocket: ExamArenaWebSocket by lazy {
        ExamArenaWebSocket(sessionManager, apiClient)
    }

    val authRepository: AuthRepository by lazy {
        AuthRepository(apiClient, sessionManager)
    }

    val matchRepository: MatchRepository by lazy {
        MatchRepository(apiClient, webSocket)
    }

    val leaderboardRepository: LeaderboardRepository by lazy {
        LeaderboardRepository(apiClient)
    }

    val practiceRepository: PracticeRepository by lazy {
        PracticeRepository(apiClient)
    }

    val userRepository: UserRepository by lazy {
        UserRepository(apiClient)
    }
}
