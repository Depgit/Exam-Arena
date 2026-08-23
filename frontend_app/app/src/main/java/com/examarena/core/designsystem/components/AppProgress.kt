package com.examarena.core.designsystem.components

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.examarena.core.designsystem.motion.MotionTokens
import com.examarena.core.designsystem.theme.ElectricCyan
import com.examarena.core.designsystem.theme.ErrorRed
import com.examarena.core.designsystem.theme.PrimaryIndigo
import com.examarena.core.designsystem.theme.SuccessGreen
import com.examarena.core.designsystem.theme.WarningAmber

@Composable
fun AppProgressBar(
    progress: Float, // 0.0f to 1.0f
    modifier: Modifier = Modifier,
    height: Dp = 8.dp,
    trackColor: Color = MaterialTheme.colorScheme.surfaceVariant,
    gradientColors: List<Color> = listOf(PrimaryIndigo, ElectricCyan)
) {
    val animatedProgress by animateFloatAsState(
        targetValue = progress.coerceIn(0f, 1f),
        animationSpec = MotionTokens.componentSpec(),
        label = "progress"
    )

    Box(
        modifier = modifier
            .fillMaxWidth()
            .height(height)
            .clip(RoundedCornerShape(height / 2))
            .background(trackColor)
    ) {
        Box(
            modifier = Modifier
                .fillMaxHeight()
                .fillMaxWidth(animatedProgress)
                .clip(RoundedCornerShape(height / 2))
                .background(Brush.horizontalGradient(gradientColors))
        )
    }
}

@Composable
fun CircularAccuracyGauge(
    accuracy: Float, // 0.0f to 100.0f
    modifier: Modifier = Modifier,
    size: Dp = 90.dp,
    strokeWidth: Dp = 8.dp
) {
    val animatedAccuracy by animateFloatAsState(
        targetValue = accuracy.coerceIn(0f, 100f),
        animationSpec = MotionTokens.componentSpec(),
        label = "accuracy"
    )

    val gaugeColor = when {
        accuracy >= 80f -> SuccessGreen
        accuracy >= 60f -> WarningAmber
        else -> ErrorRed
    }

    val trackColor = MaterialTheme.colorScheme.surfaceVariant

    Box(
        modifier = modifier.size(size),
        contentAlignment = Alignment.Center
    ) {
        Canvas(modifier = Modifier.size(size)) {
            val stroke = Stroke(width = strokeWidth.toPx(), cap = StrokeCap.Round)
            // Track
            drawArc(
                color = trackColor,
                startAngle = 135f,
                sweepAngle = 270f,
                useCenter = false,
                style = stroke
            )
            // Progress
            drawArc(
                color = gaugeColor,
                startAngle = 135f,
                sweepAngle = (animatedAccuracy / 100f) * 270f,
                useCenter = false,
                style = stroke
            )
        }

        Text(
            text = "${animatedAccuracy.toInt()}%",
            style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
            color = MaterialTheme.colorScheme.onSurface
        )
    }
}
