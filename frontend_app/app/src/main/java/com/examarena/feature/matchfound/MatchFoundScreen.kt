package com.examarena.feature.matchfound

import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.core.FastOutSlowInEasing
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.togetherWith
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
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.examarena.core.common.Resource
import com.examarena.core.designsystem.components.AppAvatar
import com.examarena.core.designsystem.theme.ElectricCyan
import com.examarena.core.designsystem.theme.ErrorRed
import com.examarena.core.designsystem.theme.PrimaryIndigo
import com.examarena.core.designsystem.theme.VividPurple
import com.examarena.data.model.MatchDetailsResponseDto
import com.examarena.data.repository.MatchRepository
import kotlinx.coroutines.delay

@Composable
fun MatchFoundScreen(
    matchId: String,
    matchRepository: MatchRepository,
    currentUserId: String?,
    onStartBattle: () -> Unit
) {
    var matchDetails by remember { mutableStateOf<MatchDetailsResponseDto?>(null) }
    var countdown by remember { mutableIntStateOf(3) }
    var isCountdownFinished by remember { mutableStateOf(false) }

    LaunchedEffect(matchId) {
        val result = matchRepository.getMatchDetails(matchId)
        if (result is Resource.Success) {
            matchDetails = result.data
        }
    }

    // 3.. 2.. 1.. GO countdown sequence
    LaunchedEffect(Unit) {
        delay(1000)
        countdown = 2
        delay(1000)
        countdown = 1
        delay(1000)
        countdown = 0 // "START!"
        delay(600)
        isCountdownFinished = true
        onStartBattle()
    }

    val players = matchDetails?.players ?: emptyList()
    val myPlayer = players.find { it.userId == currentUserId } ?: players.firstOrNull()
    val opponentPlayer = players.find { it.userId != currentUserId } ?: players.lastOrNull()

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(
                Brush.verticalGradient(
                    listOf(
                        MaterialTheme.colorScheme.background,
                        MaterialTheme.colorScheme.surface
                    )
                )
            )
            .padding(24.dp),
        contentAlignment = Alignment.Center
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Text(
                text = "MATCH FOUND!",
                style = MaterialTheme.typography.headlineLarge.copy(
                    fontWeight = FontWeight.Black,
                    letterSpacing = 2.sp
                ),
                color = ElectricCyan
            )

            Spacer(modifier = Modifier.height(40.dp))

            // VS Arena Presentation
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceEvenly,
                verticalAlignment = Alignment.CenterVertically
            ) {
                // Player Card
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    AppAvatar(
                        name = myPlayer?.username ?: "You",
                        size = 80.dp,
                        rating = myPlayer?.ratingBefore ?: 1200
                    )
                    Spacer(modifier = Modifier.height(10.dp))
                    Text(
                        text = myPlayer?.username ?: "You",
                        style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                        color = MaterialTheme.colorScheme.onBackground
                    )
                    Text(
                        text = "${myPlayer?.ratingBefore ?: 1200} ELO",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }

                // VS Badge
                Box(
                    modifier = Modifier
                        .size(54.dp)
                        .clip(CircleShape)
                        .background(
                            Brush.linearGradient(
                                listOf(ErrorRed, VividPurple)
                            )
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = "VS",
                        color = Color.White,
                        fontWeight = FontWeight.Black,
                        fontSize = 20.sp
                    )
                }

                // Opponent Card
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    AppAvatar(
                        name = opponentPlayer?.username ?: "Opponent",
                        size = 80.dp,
                        rating = opponentPlayer?.ratingBefore ?: 1200
                    )
                    Spacer(modifier = Modifier.height(10.dp))
                    Text(
                        text = opponentPlayer?.username ?: "Opponent",
                        style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                        color = MaterialTheme.colorScheme.onBackground
                    )
                    Text(
                        text = "${opponentPlayer?.ratingBefore ?: 1200} ELO",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }

            Spacer(modifier = Modifier.height(56.dp))

            // Countdown presentation
            AnimatedContent(
                targetState = countdown,
                transitionSpec = {
                    fadeIn(animationSpec = tween(200, easing = FastOutSlowInEasing)) togetherWith
                            fadeOut(animationSpec = tween(200))
                },
                label = "countdown"
            ) { targetCount ->
                if (targetCount > 0) {
                    Text(
                        text = "$targetCount",
                        style = MaterialTheme.typography.displayLarge.copy(
                            fontSize = 64.sp,
                            fontWeight = FontWeight.Black
                        ),
                        color = MaterialTheme.colorScheme.primary
                    )
                } else {
                    Surface(
                        shape = RoundedCornerShape(24.dp),
                        color = PrimaryIndigo
                    ) {
                        Text(
                            text = "START! ⚔",
                            style = MaterialTheme.typography.headlineLarge.copy(
                                fontWeight = FontWeight.Black,
                                letterSpacing = 2.sp
                            ),
                            color = Color.White,
                            modifier = Modifier.padding(horizontal = 28.dp, vertical = 12.dp)
                        )
                    }
                }
            }
        }
    }
}
