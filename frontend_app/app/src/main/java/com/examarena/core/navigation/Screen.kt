package com.examarena.core.navigation

sealed class Screen(val route: String) {
    data object Splash : Screen("splash")
    data object Onboarding : Screen("onboarding")
    data object Auth : Screen("auth")

    // Main Tabs
    data object Home : Screen("home")
    data object Leaderboard : Screen("leaderboard")
    data object PlaySetup : Screen("play_setup")
    data object PracticeSetup : Screen("practice_setup")
    data object Profile : Screen("profile")

    // Match & Battle Flow
    data object Matchmaking : Screen("matchmaking/{categoryId}/{matchType}") {
        fun createRoute(categoryId: String, matchType: String = "ranked"): String =
            "matchmaking/$categoryId/$matchType"
    }

    data object MatchFound : Screen("match_found/{matchId}") {
        fun createRoute(matchId: String): String = "match_found/$matchId"
    }

    data object Battle : Screen("battle/{matchId}") {
        fun createRoute(matchId: String): String = "battle/$matchId"
    }

    data object Result : Screen("result/{matchId}") {
        fun createRoute(matchId: String): String = "result/$matchId"
    }

    data object PracticeSession : Screen("practice_session/{sessionId}") {
        fun createRoute(sessionId: String): String = "practice_session/$sessionId"
    }
}
