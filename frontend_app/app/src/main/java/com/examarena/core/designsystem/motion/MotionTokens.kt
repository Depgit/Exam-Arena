package com.examarena.core.designsystem.motion

import androidx.compose.animation.core.CubicBezierEasing
import androidx.compose.animation.core.Easing
import androidx.compose.animation.core.FastOutSlowInEasing
import androidx.compose.animation.core.LinearOutSlowInEasing
import androidx.compose.animation.core.Spring
import androidx.compose.animation.core.spring
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInHorizontally
import androidx.compose.animation.slideOutHorizontally

object MotionTokens {
    // Durations
    const val DurationMicro = 150
    const val DurationComponent = 280
    const val DurationNavigation = 380
    const val DurationCelebration = 700

    // Easings
    val EmphasizedEasing: Easing = CubicBezierEasing(0.2f, 0.0f, 0.0f, 1.0f)
    val DecelerateEasing: Easing = LinearOutSlowInEasing
    val StandardEasing: Easing = FastOutSlowInEasing

    // Animation Specs
    fun <T> microSpec() = tween<T>(durationMillis = DurationMicro, easing = StandardEasing)
    fun <T> componentSpec() = tween<T>(durationMillis = DurationComponent, easing = EmphasizedEasing)
    fun <T> navigationSpec() = tween<T>(durationMillis = DurationNavigation, easing = EmphasizedEasing)
    fun <T> springBouncy() = spring<T>(
        dampingRatio = Spring.DampingRatioMediumBouncy,
        stiffness = Spring.StiffnessMedium
    )

    // Standard Navigation Transitions
    val NavEnterTransition = slideInHorizontally(
        initialOffsetX = { fullWidth -> (fullWidth * 0.15f).toInt() },
        animationSpec = navigationSpec()
    ) + fadeIn(animationSpec = navigationSpec())

    val NavExitTransition = slideOutHorizontally(
        targetOffsetX = { fullWidth -> (-fullWidth * 0.15f).toInt() },
        animationSpec = navigationSpec()
    ) + fadeOut(animationSpec = navigationSpec())

    val NavPopEnterTransition = slideInHorizontally(
        initialOffsetX = { fullWidth -> (-fullWidth * 0.15f).toInt() },
        animationSpec = navigationSpec()
    ) + fadeIn(animationSpec = navigationSpec())

    val NavPopExitTransition = slideOutHorizontally(
        targetOffsetX = { fullWidth -> (fullWidth * 0.15f).toInt() },
        animationSpec = navigationSpec()
    ) + fadeOut(animationSpec = navigationSpec())
}
