package com.examarena.core.navigation

import androidx.compose.animation.AnimatedContentTransitionScope
import androidx.compose.animation.core.tween
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.padding
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.navigation.NavHostController
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.navArgument
import com.examarena.core.di.AppContainer
import com.examarena.feature.auth.AuthScreen
import com.examarena.feature.auth.AuthViewModel
import com.examarena.feature.battle.BattleScreen
import com.examarena.feature.battle.BattleViewModel
import com.examarena.feature.home.HomeScreen
import com.examarena.feature.home.HomeViewModel
import com.examarena.feature.leaderboard.LeaderboardScreen
import com.examarena.feature.leaderboard.LeaderboardViewModel
import com.examarena.feature.matchfound.MatchFoundScreen
import com.examarena.feature.matchmaking.MatchmakingScreen
import com.examarena.feature.matchmaking.MatchmakingViewModel
import com.examarena.feature.onboarding.OnboardingScreen
import com.examarena.feature.play.PlaySetupScreen
import com.examarena.feature.practice.PracticeSessionScreen
import com.examarena.feature.practice.PracticeSetupScreen
import com.examarena.feature.practice.PracticeViewModel
import com.examarena.feature.profile.ProfileScreen
import com.examarena.feature.profile.ProfileViewModel
import com.examarena.feature.splash.SplashScreen
import com.examarena.feature.splash.SplashViewModel

@Composable
fun ExamArenaNavGraph(
    navController: NavHostController,
    container: AppContainer,
    currentUserId: String?,
    paddingValues: PaddingValues,
    modifier: Modifier = Modifier
) {
    NavHost(
        navController = navController,
        startDestination = Screen.Splash.route,
        modifier = modifier.padding(paddingValues),
        enterTransition = {
            slideIntoContainer(
                AnimatedContentTransitionScope.SlideDirection.Start,
                tween(300)
            )
        },
        exitTransition = {
            slideOutOfContainer(
                AnimatedContentTransitionScope.SlideDirection.Start,
                tween(300)
            )
        },
        popEnterTransition = {
            slideIntoContainer(
                AnimatedContentTransitionScope.SlideDirection.End,
                tween(300)
            )
        },
        popExitTransition = {
            slideOutOfContainer(
                AnimatedContentTransitionScope.SlideDirection.End,
                tween(300)
            )
        }
    ) {
        // ── Splash ──────────────────────────────────────────────────
        composable(Screen.Splash.route) {
            val viewModel: SplashViewModel = viewModel(
                factory = SplashViewModel.provideFactory(container.authRepository)
            )
            SplashScreen(
                viewModel = viewModel,
                onNavigateToHome = {
                    navController.navigate(Screen.Home.route) {
                        popUpTo(Screen.Splash.route) { inclusive = true }
                    }
                },
                onNavigateToOnboarding = {
                    navController.navigate(Screen.Onboarding.route) {
                        popUpTo(Screen.Splash.route) { inclusive = true }
                    }
                },
                onNavigateToAuth = {
                    navController.navigate(Screen.Auth.route) {
                        popUpTo(Screen.Splash.route) { inclusive = true }
                    }
                }
            )
        }

        // ── Onboarding ──────────────────────────────────────────────
        composable(Screen.Onboarding.route) {
            OnboardingScreen(
                onNavigateToAuth = {
                    navController.navigate(Screen.Auth.route) {
                        popUpTo(Screen.Onboarding.route) { inclusive = true }
                    }
                }
            )
        }

        // ── Auth ────────────────────────────────────────────────────
        composable(Screen.Auth.route) {
            val viewModel: AuthViewModel = viewModel(
                factory = AuthViewModel.provideFactory(container.authRepository)
            )
            AuthScreen(
                viewModel = viewModel,
                onAuthSuccess = {
                    navController.navigate(Screen.Home.route) {
                        popUpTo(Screen.Auth.route) { inclusive = true }
                    }
                }
            )
        }

        // ── Home ────────────────────────────────────────────────────
        composable(Screen.Home.route) {
            val viewModel: HomeViewModel = viewModel(
                factory = HomeViewModel.provideFactory(
                    container.authRepository,
                    container.userRepository
                )
            )
            HomeScreen(
                viewModel = viewModel,
                onStartMatchmaking = { categoryId ->
                    navController.navigate(Screen.Matchmaking.createRoute(categoryId))
                },
                onOpenPlaySetup = {
                    navController.navigate(Screen.PlaySetup.route)
                },
                onOpenPractice = { categoryId ->
                    navController.navigate(Screen.PracticeSetup.route)
                },
                onOpenLeaderboard = {
                    navController.navigate(Screen.Leaderboard.route)
                },
                onOpenProfile = {
                    navController.navigate(Screen.Profile.route)
                }
            )
        }

        // ── Play / Setup ────────────────────────────────────────────
        composable(Screen.PlaySetup.route) {
            PlaySetupScreen(
                matchRepository = container.matchRepository,
                onStartRankedMatch = { categoryId ->
                    navController.navigate(Screen.Matchmaking.createRoute(categoryId))
                },
                onNavigateToMatchFound = { matchId ->
                    navController.navigate(Screen.MatchFound.createRoute(matchId))
                },
                onBackClick = { navController.popBackStack() }
            )
        }

        // ── Matchmaking ─────────────────────────────────────────────
        composable(
            route = Screen.Matchmaking.route,
            arguments = listOf(
                navArgument("categoryId") { type = NavType.StringType },
                navArgument("matchType") { type = NavType.StringType; defaultValue = "ranked" }
            )
        ) { backStackEntry ->
            val categoryId = backStackEntry.arguments?.getString("categoryId") ?: "ssc-cgl"
            val matchType = backStackEntry.arguments?.getString("matchType") ?: "ranked"

            val viewModel: MatchmakingViewModel = viewModel(
                factory = MatchmakingViewModel.provideFactory(
                    categoryId,
                    matchType,
                    container.matchRepository
                )
            )

            MatchmakingScreen(
                viewModel = viewModel,
                onMatchFound = { matchId ->
                    navController.navigate(Screen.MatchFound.createRoute(matchId)) {
                        popUpTo(Screen.Matchmaking.route) { inclusive = true }
                    }
                },
                onCancelled = { navController.popBackStack() }
            )
        }

        // ── Match Found ─────────────────────────────────────────────
        composable(
            route = Screen.MatchFound.route,
            arguments = listOf(navArgument("matchId") { type = NavType.StringType })
        ) { backStackEntry ->
            val matchId = backStackEntry.arguments?.getString("matchId") ?: ""
            MatchFoundScreen(
                matchId = matchId,
                matchRepository = container.matchRepository,
                currentUserId = currentUserId,
                onStartBattle = {
                    navController.navigate(Screen.Battle.createRoute(matchId)) {
                        popUpTo(Screen.MatchFound.route) { inclusive = true }
                    }
                }
            )
        }

        // ── Battle ──────────────────────────────────────────────────
        composable(
            route = Screen.Battle.route,
            arguments = listOf(navArgument("matchId") { type = NavType.StringType })
        ) { backStackEntry ->
            val matchId = backStackEntry.arguments?.getString("matchId") ?: ""
            val viewModel: BattleViewModel = viewModel(
                factory = BattleViewModel.provideFactory(
                    matchId,
                    currentUserId,
                    container.matchRepository
                )
            )

            BattleScreen(
                viewModel = viewModel,
                onMatchCompleted = { completedMatchId ->
                    navController.navigate(Screen.Result.createRoute(completedMatchId)) {
                        popUpTo(Screen.Battle.route) { inclusive = true }
                    }
                }
            )
        }

        // ── Result ──────────────────────────────────────────────────
        composable(
            route = Screen.Result.route,
            arguments = listOf(navArgument("matchId") { type = NavType.StringType })
        ) { backStackEntry ->
            val matchId = backStackEntry.arguments?.getString("matchId") ?: ""
            val viewModel: com.examarena.feature.result.ResultViewModel = viewModel(
                factory = com.examarena.feature.result.ResultViewModel.provideFactory(
                    matchId,
                    currentUserId,
                    container.matchRepository
                )
            )

            com.examarena.feature.result.ResultScreen(
                viewModel = viewModel,
                onPlayAgain = {
                    navController.navigate(Screen.PlaySetup.route) {
                        popUpTo(Screen.Result.route) { inclusive = true }
                    }
                },
                onOpenLeaderboard = {
                    navController.navigate(Screen.Leaderboard.route)
                },
                onReturnHome = {
                    navController.navigate(Screen.Home.route) {
                        popUpTo(Screen.Result.route) { inclusive = true }
                    }
                }
            )
        }

        // ── Leaderboard ─────────────────────────────────────────────
        composable(Screen.Leaderboard.route) {
            val viewModel: LeaderboardViewModel = viewModel(
                factory = LeaderboardViewModel.provideFactory(container.leaderboardRepository)
            )
            LeaderboardScreen(
                viewModel = viewModel,
                onBackClick = { navController.popBackStack() }
            )
        }

        // ── Practice Setup ──────────────────────────────────────────
        composable(Screen.PracticeSetup.route) {
            val viewModel: PracticeViewModel = viewModel(
                factory = PracticeViewModel.provideFactory(container.practiceRepository)
            )
            PracticeSetupScreen(
                viewModel = viewModel,
                onStartSession = {
                    val sessionId = viewModel.uiState.value.session?.id ?: ""
                    navController.navigate(Screen.PracticeSession.createRoute(sessionId))
                },
                onBackClick = { navController.popBackStack() }
            )
        }

        // ── Practice Session ────────────────────────────────────────
        composable(
            route = Screen.PracticeSession.route,
            arguments = listOf(navArgument("sessionId") { type = NavType.StringType })
        ) {
            val viewModel: PracticeViewModel = viewModel(
                factory = PracticeViewModel.provideFactory(container.practiceRepository)
            )
            PracticeSessionScreen(
                viewModel = viewModel,
                onBackClick = {
                    navController.navigate(Screen.Home.route) {
                        popUpTo(Screen.PracticeSession.route) { inclusive = true }
                    }
                }
            )
        }

        // ── Profile ─────────────────────────────────────────────────
        composable(Screen.Profile.route) {
            val viewModel: ProfileViewModel = viewModel(
                factory = ProfileViewModel.provideFactory(
                    container.authRepository,
                    container.userRepository,
                    container.sessionManager
                )
            )
            ProfileScreen(
                viewModel = viewModel,
                onLoggedOut = {
                    navController.navigate(Screen.Auth.route) {
                        popUpTo(0) { inclusive = true }
                    }
                },
                onBackClick = { navController.popBackStack() }
            )
        }
    }
}
