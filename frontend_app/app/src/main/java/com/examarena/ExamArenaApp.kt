package com.examarena

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import com.examarena.core.designsystem.components.AppBottomBar
import com.examarena.core.designsystem.theme.ExamArenaTheme
import com.examarena.core.di.AppContainer
import com.examarena.core.navigation.ExamArenaNavGraph
import com.examarena.core.navigation.Screen

@Composable
fun ExamArenaApp(container: AppContainer) {
    val navController = rememberNavController()
    val navBackStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = navBackStackEntry?.destination?.route ?: Screen.Splash.route

    val currentUserId by container.sessionManager.userIdFlow.collectAsStateWithLifecycle(initialValue = null)
    val darkThemeSetting by container.sessionManager.isDarkThemeFlow.collectAsStateWithLifecycle(initialValue = null)

    val isDarkTheme = darkThemeSetting ?: isSystemInDarkTheme()

    val topLevelRoutes = setOf(
        Screen.Home.route,
        Screen.Leaderboard.route,
        Screen.PlaySetup.route,
        Screen.PracticeSetup.route,
        Screen.Profile.route
    )

    val showBottomBar = currentRoute in topLevelRoutes

    ExamArenaTheme(darkTheme = isDarkTheme) {
        Scaffold(
            bottomBar = {
                if (showBottomBar) {
                    AppBottomBar(
                        currentRoute = currentRoute,
                        onNavigate = { targetRoute ->
                            if (currentRoute != targetRoute) {
                                navController.navigate(targetRoute) {
                                    popUpTo(Screen.Home.route) {
                                        saveState = true
                                    }
                                    launchSingleTop = true
                                    restoreState = true
                                }
                            }
                        }
                    )
                }
            }
        ) { paddingValues ->
            ExamArenaNavGraph(
                navController = navController,
                container = container,
                currentUserId = currentUserId,
                paddingValues = paddingValues
            )
        }
    }
}
