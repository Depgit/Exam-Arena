package com.examarena.feature.play

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
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
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.FlashOn
import androidx.compose.material.icons.filled.Group
import androidx.compose.material.icons.filled.MeetingRoom
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.examarena.core.common.Resource
import com.examarena.core.designsystem.components.AppButton
import com.examarena.core.designsystem.components.AppCard
import com.examarena.core.designsystem.components.AppTextField
import com.examarena.core.designsystem.components.AppTopBar
import com.examarena.core.designsystem.components.ButtonVariant
import com.examarena.core.designsystem.theme.ElectricCyan
import com.examarena.core.designsystem.theme.PrimaryIndigo
import com.examarena.core.designsystem.theme.VividPurple
import com.examarena.data.model.ExamCategory
import com.examarena.data.repository.MatchRepository
import kotlinx.coroutines.launch

@Composable
fun PlaySetupScreen(
    matchRepository: MatchRepository,
    onStartRankedMatch: (categoryId: String) -> Unit,
    onNavigateToMatchFound: (matchId: String) -> Unit,
    onBackClick: () -> Unit
) {
    var selectedCategory by remember { mutableStateOf(ExamCategory.SSCCGL) }
    var roomCodeInput by remember { mutableStateOf("") }
    var createdRoomCode by remember { mutableStateOf<String?>(null) }
    var createdMatchId by remember { mutableStateOf<String?>(null) }
    var isLoading by remember { mutableStateOf(false) }
    var errorMessage by remember { mutableStateOf<String?>(null) }

    val scope = rememberCoroutineScope()
    val scrollState = rememberScrollState()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        AppTopBar(
            title = "Play Arena",
            onBackClick = onBackClick
        )

        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(scrollState)
                .padding(horizontal = 20.dp, vertical = 16.dp)
        ) {
            // ── 1. Select Category ──────────────────────────────────
            Text(
                text = "1. Select Category",
                style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                color = MaterialTheme.colorScheme.onBackground
            )

            Spacer(modifier = Modifier.height(10.dp))

            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                ExamCategory.entries.chunked(2).forEach { rowCategories ->
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.spacedBy(10.dp)
                    ) {
                        rowCategories.forEach { category ->
                            val isSelected = selectedCategory == category
                            Surface(
                                modifier = Modifier
                                    .weight(1f)
                                    .clip(RoundedCornerShape(12.dp))
                                    .clickable { selectedCategory = category },
                                shape = RoundedCornerShape(12.dp),
                                color = if (isSelected) PrimaryIndigo.copy(alpha = 0.15f) else MaterialTheme.colorScheme.surface,
                                border = if (isSelected) androidx.compose.foundation.BorderStroke(2.dp, PrimaryIndigo) else null,
                                tonalElevation = 2.dp
                            ) {
                                Row(
                                    modifier = Modifier.padding(14.dp),
                                    verticalAlignment = Alignment.CenterVertically
                                ) {
                                    Text(text = category.iconEmoji, fontSize = 20.sp)
                                    Spacer(modifier = Modifier.width(8.dp))
                                    Text(
                                        text = category.title,
                                        style = MaterialTheme.typography.labelLarge.copy(
                                            fontWeight = if (isSelected) FontWeight.Bold else FontWeight.Medium
                                        ),
                                        color = if (isSelected) PrimaryIndigo else MaterialTheme.colorScheme.onSurface
                                    )
                                }
                            }
                        }
                    }
                }
            }

            Spacer(modifier = Modifier.height(28.dp))

            // ── 2. 1v1 Ranked Matchmaking ───────────────────────────
            AppCard(
                modifier = Modifier.fillMaxWidth(),
                containerColor = MaterialTheme.colorScheme.surface,
                elevation = 2.dp
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Box(
                        modifier = Modifier
                            .size(40.dp)
                            .clip(RoundedCornerShape(10.dp))
                            .background(PrimaryIndigo.copy(alpha = 0.12f)),
                        contentAlignment = Alignment.Center
                    ) {
                        androidx.compose.material3.Icon(
                            imageVector = Icons.Default.FlashOn,
                            contentDescription = null,
                            tint = PrimaryIndigo
                        )
                    }
                    Spacer(modifier = Modifier.width(12.dp))
                    Column {
                        Text(
                            text = "Ranked 1v1 Matchmaking",
                            style = MaterialTheme.typography.titleSmall.copy(fontWeight = FontWeight.Bold),
                            color = MaterialTheme.colorScheme.onSurface
                        )
                        Text(
                            text = "Matched against players of equal ELO",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                }

                Spacer(modifier = Modifier.height(16.dp))

                AppButton(
                    text = "Find Opponent (${selectedCategory.code})",
                    onClick = { onStartRankedMatch(selectedCategory.id) }
                )
            }

            Spacer(modifier = Modifier.height(24.dp))

            // ── 3. Friendly Room Match ──────────────────────────────
            AppCard(
                modifier = Modifier.fillMaxWidth(),
                containerColor = MaterialTheme.colorScheme.surface,
                elevation = 2.dp
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Box(
                        modifier = Modifier
                            .size(40.dp)
                            .clip(RoundedCornerShape(10.dp))
                            .background(VividPurple.copy(alpha = 0.12f)),
                        contentAlignment = Alignment.Center
                    ) {
                        androidx.compose.material3.Icon(
                            imageVector = Icons.Default.Group,
                            contentDescription = null,
                            tint = VividPurple
                        )
                    }
                    Spacer(modifier = Modifier.width(12.dp))
                    Column {
                        Text(
                            text = "Friendly Battle Room",
                            style = MaterialTheme.typography.titleSmall.copy(fontWeight = FontWeight.Bold),
                            color = MaterialTheme.colorScheme.onSurface
                        )
                        Text(
                            text = "Create or join custom rooms with a 6-digit code",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                }

                Spacer(modifier = Modifier.height(16.dp))

                if (createdRoomCode != null) {
                    Surface(
                        modifier = Modifier.fillMaxWidth(),
                        shape = RoundedCornerShape(12.dp),
                        color = VividPurple.copy(alpha = 0.12f)
                    ) {
                        Column(
                            modifier = Modifier.padding(16.dp),
                            horizontalAlignment = Alignment.CenterHorizontally
                        ) {
                            Text(
                                text = "ROOM CODE",
                                style = MaterialTheme.typography.labelSmall.copy(
                                    fontWeight = FontWeight.Bold,
                                    letterSpacing = 1.5.sp
                                ),
                                color = VividPurple
                            )
                            Spacer(modifier = Modifier.height(4.dp))
                            Text(
                                text = createdRoomCode ?: "",
                                style = MaterialTheme.typography.displayMedium.copy(
                                    fontWeight = FontWeight.Black,
                                    letterSpacing = 4.sp
                                ),
                                color = VividPurple
                            )
                            Spacer(modifier = Modifier.height(6.dp))
                            Text(
                                text = "Share this code with your friend. The match starts when they join!",
                                style = MaterialTheme.typography.bodySmall,
                                textAlign = TextAlign.Center,
                                color = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }
                    }
                } else {
                    AppButton(
                        text = "Create Room Code",
                        onClick = {
                            scope.launch {
                                isLoading = true
                                errorMessage = null
                                matchRepository.connectWebSocket()
                                val result = matchRepository.createFriendMatch(selectedCategory.id)
                                isLoading = false
                                if (result is Resource.Success) {
                                    createdRoomCode = result.data.roomCode
                                    createdMatchId = result.data.matchId
                                } else if (result is Resource.Error) {
                                    errorMessage = result.message
                                }
                            }
                        },
                        variant = ButtonVariant.Secondary,
                        loading = isLoading
                    )
                }

                Spacer(modifier = Modifier.height(20.dp))

                // Join Room Input
                Text(
                    text = "Or Join Existing Room:",
                    style = MaterialTheme.typography.labelMedium.copy(fontWeight = FontWeight.SemiBold),
                    color = MaterialTheme.colorScheme.onSurface
                )

                Spacer(modifier = Modifier.height(8.dp))

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    AppTextField(
                        value = roomCodeInput,
                        onValueChange = { roomCodeInput = it.uppercase().take(6) },
                        label = "6-Digit Code",
                        modifier = Modifier.weight(1f)
                    )
                    Spacer(modifier = Modifier.width(10.dp))
                    AppButton(
                        text = "Join",
                        onClick = {
                            if (roomCodeInput.length == 6) {
                                scope.launch {
                                    isLoading = true
                                    errorMessage = null
                                    matchRepository.connectWebSocket()
                                    val result = matchRepository.joinFriendMatch(roomCodeInput.trim())
                                    isLoading = false
                                    if (result is Resource.Success) {
                                        onNavigateToMatchFound(result.data.matchId)
                                    } else if (result is Resource.Error) {
                                        errorMessage = result.message
                                    }
                                }
                            }
                        },
                        fillWidth = false,
                        enabled = roomCodeInput.length == 6 && !isLoading
                    )
                }

                if (errorMessage != null) {
                    Spacer(modifier = Modifier.height(8.dp))
                    Text(
                        text = errorMessage ?: "",
                        color = MaterialTheme.colorScheme.error,
                        style = MaterialTheme.typography.bodySmall
                    )
                }
            }

            Spacer(modifier = Modifier.height(80.dp))
        }
    }
}
