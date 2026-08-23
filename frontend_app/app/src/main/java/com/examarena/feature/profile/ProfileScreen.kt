package com.examarena.feature.profile

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
import androidx.compose.material.icons.filled.DarkMode
import androidx.compose.material.icons.filled.Dns
import androidx.compose.material.icons.filled.EmojiEvents
import androidx.compose.material.icons.filled.FlashOn
import androidx.compose.material.icons.filled.Insights
import androidx.compose.material.icons.filled.Logout
import androidx.compose.material.icons.filled.VolumeUp
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.examarena.core.designsystem.components.AppAvatar
import com.examarena.core.designsystem.components.AppBadge
import com.examarena.core.designsystem.components.AppButton
import com.examarena.core.designsystem.components.AppCard
import com.examarena.core.designsystem.components.AppTextField
import com.examarena.core.designsystem.components.AppTopBar
import com.examarena.core.designsystem.components.BadgeVariant
import com.examarena.core.designsystem.components.ButtonVariant
import com.examarena.core.designsystem.theme.BronzeRank
import com.examarena.core.designsystem.theme.ErrorRed
import com.examarena.core.designsystem.theme.GoldWinner
import com.examarena.core.designsystem.theme.PrimaryIndigo
import com.examarena.core.designsystem.theme.SilverRank
import com.examarena.core.designsystem.theme.SuccessGreen
import com.examarena.core.designsystem.theme.VividPurple

@Composable
fun ProfileScreen(
    viewModel: ProfileViewModel,
    onLoggedOut: () -> Unit,
    onBackClick: (() -> Unit)? = null
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val apiUrl by viewModel.apiUrlFlow.collectAsStateWithLifecycle()
    val wsUrl by viewModel.wsUrlFlow.collectAsStateWithLifecycle()
    val isDarkTheme by viewModel.isDarkThemeFlow.collectAsStateWithLifecycle()
    val isSoundEnabled by viewModel.isSoundEnabledFlow.collectAsStateWithLifecycle()

    val scrollState = rememberScrollState()
    var showServerConfigDialog by remember { mutableStateOf(false) }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        AppTopBar(
            title = "Profile & Settings",
            onBackClick = onBackClick,
            centerAligned = true
        )

        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(scrollState)
                .padding(horizontal = 20.dp, vertical = 12.dp)
        ) {
            // ── User Header Card ────────────────────────────────────
            AppCard(
                modifier = Modifier.fillMaxWidth(),
                containerColor = MaterialTheme.colorScheme.surface,
                elevation = 2.dp,
                contentPadding = 20.dp
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    AppAvatar(
                        name = uiState.user?.username ?: "Player",
                        size = 64.dp
                    )

                    Spacer(modifier = Modifier.width(16.dp))

                    Column {
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Text(
                                text = uiState.user?.username ?: "Player",
                                style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                                color = MaterialTheme.colorScheme.onSurface
                            )
                            Spacer(modifier = Modifier.width(8.dp))
                            AppBadge(
                                text = (uiState.user?.role ?: "USER").uppercase(),
                                variant = BadgeVariant.Primary
                            )
                        }

                        Spacer(modifier = Modifier.height(4.dp))

                        Text(
                            text = uiState.user?.email ?: "competitor@examarena.local",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.height(20.dp))

            // ── Lifetime Statistics ─────────────────────────────────
            Text(
                text = "Lifetime Arena Statistics",
                style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                color = MaterialTheme.colorScheme.onBackground
            )

            Spacer(modifier = Modifier.height(10.dp))

            AppCard(
                modifier = Modifier.fillMaxWidth(),
                containerColor = MaterialTheme.colorScheme.surface,
                elevation = 2.dp,
                contentPadding = 16.dp
            ) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceAround
                ) {
                    ProfileStatColumn("Matches", "${uiState.statistics?.totalMatches ?: 0}", GoldWinner)
                    ProfileStatColumn("Wins", "${uiState.statistics?.wins ?: 0}", SuccessGreen)
                    ProfileStatColumn("Win Streak", "${uiState.statistics?.currentWinStreak ?: 0}🔥", VividPurple)
                    ProfileStatColumn("Questions", "${uiState.statistics?.totalQuestionsSolved ?: 0}", PrimaryIndigo)
                }
            }

            Spacer(modifier = Modifier.height(20.dp))

            // ── Category Ratings Breakdown ──────────────────────────
            if (uiState.ratings.isNotEmpty()) {
                Text(
                    text = "Category Ratings",
                    style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                    color = MaterialTheme.colorScheme.onBackground
                )

                Spacer(modifier = Modifier.height(10.dp))

                AppCard(
                    modifier = Modifier.fillMaxWidth(),
                    containerColor = MaterialTheme.colorScheme.surface,
                    elevation = 1.dp
                ) {
                    uiState.ratings.forEach { ratingDto ->
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(vertical = 8.dp),
                            horizontalArrangement = Arrangement.SpaceBetween,
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Text(
                                text = ratingDto.examCategoryId.uppercase(),
                                style = MaterialTheme.typography.bodyMedium.copy(fontWeight = FontWeight.SemiBold),
                                color = MaterialTheme.colorScheme.onSurface
                            )
                            Text(
                                text = "${ratingDto.rating} ELO",
                                style = MaterialTheme.typography.labelLarge.copy(fontWeight = FontWeight.Black),
                                color = PrimaryIndigo
                            )
                        }
                    }
                }

                Spacer(modifier = Modifier.height(20.dp))
            }

            // ── Preferences & Settings ──────────────────────────────
            Text(
                text = "Preferences",
                style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                color = MaterialTheme.colorScheme.onBackground
            )

            Spacer(modifier = Modifier.height(10.dp))

            AppCard(
                modifier = Modifier.fillMaxWidth(),
                containerColor = MaterialTheme.colorScheme.surface,
                elevation = 1.dp
            ) {
                // Sound Setting
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Icon(
                            imageVector = Icons.Default.VolumeUp,
                            contentDescription = null,
                            tint = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                        Spacer(modifier = Modifier.width(12.dp))
                        Text(
                            text = "Sound & Effects",
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.onSurface
                        )
                    }
                    Switch(
                        checked = isSoundEnabled,
                        onCheckedChange = viewModel::toggleSound
                    )
                }

                Spacer(modifier = Modifier.height(8.dp))

                // Server Config Entry
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .clip(RoundedCornerShape(8.dp))
                        .clickable { showServerConfigDialog = true }
                        .padding(vertical = 8.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Icon(
                            imageVector = Icons.Default.Dns,
                            contentDescription = null,
                            tint = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                        Spacer(modifier = Modifier.width(12.dp))
                        Column {
                            Text(
                                text = "Server Network Endpoint",
                                style = MaterialTheme.typography.bodyMedium,
                                color = MaterialTheme.colorScheme.onSurface
                            )
                            Text(
                                text = apiUrl,
                                style = MaterialTheme.typography.bodySmall,
                                color = MaterialTheme.colorScheme.primary
                            )
                        }
                    }
                }
            }

            Spacer(modifier = Modifier.height(28.dp))

            // Logout Button
            AppButton(
                text = "Log Out of Arena",
                onClick = { viewModel.logout(onLoggedOut) },
                variant = ButtonVariant.Danger,
                leadingIcon = Icons.Default.Logout
            )

            Spacer(modifier = Modifier.height(80.dp)) // padding for bottom nav
        }
    }

    // Server Configuration Dialog
    if (showServerConfigDialog) {
        var tempApiUrl by remember { mutableStateOf(apiUrl) }
        var tempWsUrl by remember { mutableStateOf(wsUrl) }

        AlertDialog(
            onDismissRequest = { showServerConfigDialog = false },
            title = {
                Text(
                    text = "Backend Server Configuration",
                    style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold)
                )
            },
            text = {
                Column {
                    Text(
                        text = "Customize host URLs to connect with local or staging backend:",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                    Spacer(modifier = Modifier.height(12.dp))
                    AppTextField(
                        value = tempApiUrl,
                        onValueChange = { tempApiUrl = it },
                        label = "HTTP API Base URL"
                    )
                    Spacer(modifier = Modifier.height(10.dp))
                    AppTextField(
                        value = tempWsUrl,
                        onValueChange = { tempWsUrl = it },
                        label = "WebSocket URL"
                    )
                }
            },
            confirmButton = {
                TextButton(onClick = {
                    viewModel.updateServerConfig(tempApiUrl, tempWsUrl)
                    showServerConfigDialog = false
                }) {
                    Text("Save & Apply", fontWeight = FontWeight.Bold)
                }
            },
            dismissButton = {
                TextButton(onClick = { showServerConfigDialog = false }) {
                    Text("Cancel")
                }
            }
        )
    }
}

@Composable
private fun ProfileStatColumn(label: String, value: String, color: Color) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Text(
            text = value,
            style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Black),
            color = color
        )
        Text(
            text = label,
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}
