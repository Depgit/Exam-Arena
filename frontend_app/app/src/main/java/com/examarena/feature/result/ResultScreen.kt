package com.examarena.feature.result

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.EmojiEvents
import androidx.compose.material.icons.filled.MilitaryTech
import androidx.compose.material.icons.filled.TrendingDown
import androidx.compose.material.icons.filled.TrendingUp
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.examarena.core.designsystem.components.AppAvatar
import com.examarena.core.designsystem.components.AppButton
import com.examarena.core.designsystem.components.AppCard
import com.examarena.core.designsystem.components.ButtonVariant
import com.examarena.core.designsystem.theme.ElectricCyan
import com.examarena.core.designsystem.theme.ErrorRed
import com.examarena.core.designsystem.theme.GoldWinner
import com.examarena.core.designsystem.theme.PrimaryIndigo
import com.examarena.core.designsystem.theme.SuccessGreen
import com.examarena.core.designsystem.theme.VividPurple

@Composable
fun ResultScreen(
    viewModel: ResultViewModel,
    onPlayAgain: () -> Unit,
    onOpenLeaderboard: () -> Unit,
    onReturnHome: () -> Unit
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val scrollState = rememberScrollState()

    if (uiState.isLoading) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(MaterialTheme.colorScheme.background),
            contentAlignment = Alignment.Center
        ) {
            CircularProgressIndicator(color = MaterialTheme.colorScheme.primary)
        }
        return
    }

    val isWin = uiState.isWinner
    val primaryHeaderColor = if (isWin) GoldWinner else MaterialTheme.colorScheme.primary

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
            .verticalScroll(scrollState)
            .padding(horizontal = 20.dp, vertical = 32.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        // ── Trophy / Result Icon ────────────────────────────────────
        Box(
            modifier = Modifier
                .size(90.dp)
                .clip(CircleShape)
                .background(
                    Brush.linearGradient(
                        if (isWin) listOf(GoldWinner, Color(0xFFFF8C00))
                        else listOf(PrimaryIndigo, VividPurple)
                    )
                ),
            contentAlignment = Alignment.Center
        ) {
            Icon(
                imageVector = if (isWin) Icons.Default.EmojiEvents else Icons.Default.MilitaryTech,
                contentDescription = null,
                modifier = Modifier.size(52.dp),
                tint = Color.White
            )
        }

        Spacer(modifier = Modifier.height(16.dp))

        Text(
            text = if (isWin) "YOU WIN! 🏆" else if (uiState.isDraw) "DRAW MATCH ⚔" else "GOOD GAME 👏",
            style = MaterialTheme.typography.headlineLarge.copy(
                fontWeight = FontWeight.Black,
                letterSpacing = 1.sp
            ),
            color = MaterialTheme.colorScheme.onBackground
        )

        Spacer(modifier = Modifier.height(6.dp))

        // Rating Delta Pill
        Surface(
            shape = RoundedCornerShape(20.dp),
            color = if (uiState.ratingDelta >= 0) SuccessGreen.copy(alpha = 0.15f) else ErrorRed.copy(alpha = 0.15f)
        ) {
            Row(
                verticalAlignment = Alignment.CenterVertically,
                modifier = Modifier.padding(horizontal = 14.dp, vertical = 6.dp)
            ) {
                Icon(
                    imageVector = if (uiState.ratingDelta >= 0) Icons.Default.TrendingUp else Icons.Default.TrendingDown,
                    contentDescription = null,
                    tint = if (uiState.ratingDelta >= 0) SuccessGreen else ErrorRed,
                    modifier = Modifier.size(16.dp)
                )
                Spacer(modifier = Modifier.width(6.dp))
                Text(
                    text = "${if (uiState.ratingDelta >= 0) "+" else ""}${uiState.ratingDelta} Arena Rating",
                    style = MaterialTheme.typography.labelLarge.copy(fontWeight = FontWeight.Bold),
                    color = if (uiState.ratingDelta >= 0) SuccessGreen else ErrorRed
                )
            }
        }

        Spacer(modifier = Modifier.height(28.dp))

        // ── Scores Card ─────────────────────────────────────────────
        AppCard(
            modifier = Modifier.fillMaxWidth(),
            containerColor = MaterialTheme.colorScheme.surface,
            elevation = 3.dp,
            contentPadding = 20.dp
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceAround,
                verticalAlignment = Alignment.CenterVertically
            ) {
                // You
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    AppAvatar(
                        name = uiState.myPlayer?.username ?: "You",
                        size = 56.dp,
                        rating = uiState.myPlayer?.ratingAfter ?: 1200
                    )
                    Spacer(modifier = Modifier.height(6.dp))
                    Text(
                        text = "YOU",
                        style = MaterialTheme.typography.labelSmall.copy(fontWeight = FontWeight.Bold),
                        color = PrimaryIndigo
                    )
                    Text(
                        text = "${uiState.myPlayer?.score ?: 0}",
                        style = MaterialTheme.typography.displayMedium.copy(fontWeight = FontWeight.Black),
                        color = MaterialTheme.colorScheme.onSurface
                    )
                }

                Text(
                    text = "VS",
                    style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Black),
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )

                // Opponent
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    AppAvatar(
                        name = uiState.opponentPlayer?.username ?: "Opponent",
                        size = 56.dp,
                        rating = uiState.opponentPlayer?.ratingAfter ?: 1200
                    )
                    Spacer(modifier = Modifier.height(6.dp))
                    Text(
                        text = (uiState.opponentPlayer?.username ?: "Opponent").uppercase(),
                        style = MaterialTheme.typography.labelSmall.copy(fontWeight = FontWeight.Bold),
                        color = VividPurple
                    )
                    Text(
                        text = "${uiState.opponentPlayer?.score ?: 0}",
                        style = MaterialTheme.typography.displayMedium.copy(fontWeight = FontWeight.Black),
                        color = MaterialTheme.colorScheme.onSurface
                    )
                }
            }
        }

        Spacer(modifier = Modifier.height(20.dp))

        // ── Match Breakdown Card ────────────────────────────────────
        AppCard(
            modifier = Modifier.fillMaxWidth(),
            containerColor = MaterialTheme.colorScheme.surface,
            elevation = 2.dp,
            contentPadding = 16.dp
        ) {
            Text(
                text = "PERFORMANCE BREAKDOWN",
                style = MaterialTheme.typography.labelSmall.copy(
                    fontWeight = FontWeight.Bold,
                    letterSpacing = 1.sp
                ),
                color = MaterialTheme.colorScheme.primary
            )

            Spacer(modifier = Modifier.height(14.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text(
                    text = "Accuracy",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurface
                )
                Text(
                    text = "${uiState.accuracyPercent}%",
                    style = MaterialTheme.typography.bodyMedium.copy(fontWeight = FontWeight.Bold),
                    color = if (uiState.accuracyPercent >= 70) SuccessGreen else MaterialTheme.colorScheme.onSurface
                )
            }

            Spacer(modifier = Modifier.height(10.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text(
                    text = "Correct Answers",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurface
                )
                Text(
                    text = "${uiState.correctCount} / ${uiState.totalQuestions}",
                    style = MaterialTheme.typography.bodyMedium.copy(fontWeight = FontWeight.Bold),
                    color = MaterialTheme.colorScheme.onSurface
                )
            }
        }

        Spacer(modifier = Modifier.height(32.dp))

        // ── Action Buttons ──────────────────────────────────────────
        AppButton(
            text = "⚔ Play Again",
            onClick = onPlayAgain
        )

        Spacer(modifier = Modifier.height(12.dp))

        AppButton(
            text = "🏆 View Rankings",
            onClick = onOpenLeaderboard,
            variant = ButtonVariant.Outlined
        )

        Spacer(modifier = Modifier.height(12.dp))

        AppButton(
            text = "Return Home",
            onClick = onReturnHome,
            variant = ButtonVariant.Outlined
        )
    }
}
