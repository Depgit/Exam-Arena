package com.examarena.feature.leaderboard

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
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.EmojiEvents
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.examarena.core.designsystem.components.AppAvatar
import com.examarena.core.designsystem.components.AppCard
import com.examarena.core.designsystem.components.AppEmptyState
import com.examarena.core.designsystem.components.AppErrorState
import com.examarena.core.designsystem.components.AppTopBar
import com.examarena.core.designsystem.theme.BronzeRank
import com.examarena.core.designsystem.theme.GoldWinner
import com.examarena.core.designsystem.theme.PrimaryIndigo
import com.examarena.core.designsystem.theme.SilverRank
import com.examarena.data.model.ExamCategory
import com.examarena.data.model.LeaderboardEntryDto

@Composable
fun LeaderboardScreen(
    viewModel: LeaderboardViewModel,
    onBackClick: (() -> Unit)? = null
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        AppTopBar(
            title = "Leaderboards",
            onBackClick = onBackClick,
            centerAligned = true
        )

        // Category Horizontal Tabs
        LazyRow(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp, vertical = 8.dp),
            horizontalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            items(ExamCategory.entries) { category ->
                val isSelected = uiState.selectedCategory == category
                Surface(
                    shape = RoundedCornerShape(20.dp),
                    color = if (isSelected) PrimaryIndigo else MaterialTheme.colorScheme.surfaceVariant,
                    modifier = Modifier
                        .clip(RoundedCornerShape(20.dp))
                        .clickable { viewModel.selectCategory(category) }
                ) {
                    Row(
                        modifier = Modifier.padding(horizontal = 14.dp, vertical = 8.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text(text = category.iconEmoji, fontSize = 14.sp)
                        Spacer(modifier = Modifier.width(6.dp))
                        Text(
                            text = category.code,
                            style = MaterialTheme.typography.labelMedium.copy(
                                fontWeight = if (isSelected) FontWeight.Bold else FontWeight.Medium
                            ),
                            color = if (isSelected) Color.White else MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                }
            }
        }

        Spacer(modifier = Modifier.height(8.dp))

        if (uiState.isLoading) {
            Box(
                modifier = Modifier.fillMaxSize(),
                contentAlignment = Alignment.Center
            ) {
                CircularProgressIndicator(color = MaterialTheme.colorScheme.primary)
            }
        } else if (uiState.errorMessage != null) {
            AppErrorState(
                message = uiState.errorMessage ?: "Failed to load rankings",
                onRetry = { viewModel.loadLeaderboard() }
            )
        } else if (uiState.entries.isEmpty()) {
            AppEmptyState(
                title = "No Rankings Yet",
                subtitle = "Be the first to battle and conquer this category!",
                icon = Icons.Default.EmojiEvents
            )
        } else {
            val top3 = uiState.entries.take(3)
            val rest = uiState.entries.drop(3)

            LazyColumn(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(horizontal = 16.dp)
            ) {
                // ── Top 3 Podium Section ────────────────────────────
                item {
                    PodiumView(top3)
                    Spacer(modifier = Modifier.height(24.dp))
                }

                // ── Ranks 4+ List ───────────────────────────────────
                itemsIndexed(rest) { index, entry ->
                    RankRowItem(rank = index + 4, entry = entry)
                    Spacer(modifier = Modifier.height(10.dp))
                }

                item {
                    Spacer(modifier = Modifier.height(80.dp)) // padding for bottom nav
                }
            }
        }
    }
}

@Composable
private fun PodiumView(top3: List<LeaderboardEntryDto>) {
    val rank1 = top3.getOrNull(0)
    val rank2 = top3.getOrNull(1)
    val rank3 = top3.getOrNull(2)

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(top = 16.dp),
        horizontalArrangement = Arrangement.SpaceEvenly,
        verticalAlignment = Alignment.Bottom
    ) {
        // Rank 2 (Silver)
        if (rank2 != null) {
            PodiumStep(
                entry = rank2,
                rankNumber = 2,
                badgeColor = SilverRank,
                avatarSize = 64.dp,
                pedestalHeight = 90.dp
            )
        } else {
            Spacer(modifier = Modifier.weight(1f))
        }

        // Rank 1 (Gold - Elevated in Center)
        if (rank1 != null) {
            PodiumStep(
                entry = rank1,
                rankNumber = 1,
                badgeColor = GoldWinner,
                avatarSize = 78.dp,
                pedestalHeight = 120.dp
            )
        }

        // Rank 3 (Bronze)
        if (rank3 != null) {
            PodiumStep(
                entry = rank3,
                rankNumber = 3,
                badgeColor = BronzeRank,
                avatarSize = 60.dp,
                pedestalHeight = 70.dp
            )
        } else {
            Spacer(modifier = Modifier.weight(1f))
        }
    }
}

@Composable
private fun PodiumStep(
    entry: LeaderboardEntryDto,
    rankNumber: Int,
    badgeColor: Color,
    avatarSize: androidx.compose.ui.unit.Dp,
    pedestalHeight: androidx.compose.ui.unit.Dp
) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = Modifier.width(100.dp)
    ) {
        // Avatar + Badge
        Box(contentAlignment = Alignment.BottomCenter) {
            AppAvatar(
                name = entry.username,
                size = avatarSize,
                rating = entry.rating
            )
            Surface(
                shape = CircleShape,
                color = badgeColor,
                modifier = Modifier.padding(bottom = (-6).dp)
            ) {
                Text(
                    text = "$rankNumber",
                    color = Color.White,
                    fontWeight = FontWeight.Black,
                    fontSize = 12.sp,
                    modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp)
                )
            }
        }

        Spacer(modifier = Modifier.height(10.dp))

        Text(
            text = entry.username,
            style = MaterialTheme.typography.titleSmall.copy(fontWeight = FontWeight.Bold),
            color = MaterialTheme.colorScheme.onSurface,
            maxLines = 1
        )

        Text(
            text = "${entry.rating} ELO",
            style = MaterialTheme.typography.bodySmall.copy(fontWeight = FontWeight.SemiBold),
            color = MaterialTheme.colorScheme.primary
        )

        Spacer(modifier = Modifier.height(8.dp))

        // Pedestal Box
        Surface(
            modifier = Modifier
                .fillMaxWidth()
                .height(pedestalHeight),
            shape = RoundedCornerShape(topStart = 12.dp, topEnd = 12.dp),
            color = MaterialTheme.colorScheme.surfaceVariant
        ) {
            Box(contentAlignment = Alignment.Center) {
                Text(
                    text = when (rankNumber) {
                        1 -> "🥇 1ST"
                        2 -> "🥈 2ND"
                        else -> "🥉 3RD"
                    },
                    style = MaterialTheme.typography.labelSmall.copy(fontWeight = FontWeight.Black),
                    color = badgeColor
                )
            }
        }
    }
}

@Composable
private fun RankRowItem(rank: Int, entry: LeaderboardEntryDto) {
    AppCard(
        modifier = Modifier.fillMaxWidth(),
        containerColor = MaterialTheme.colorScheme.surface,
        elevation = 1.dp,
        contentPadding = 12.dp
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically
        ) {
            // Rank Number
            Text(
                text = "#$rank",
                style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Black),
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.width(36.dp)
            )

            // Avatar
            AppAvatar(name = entry.username, size = 42.dp, rating = entry.rating)

            Spacer(modifier = Modifier.width(12.dp))

            // Username
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = entry.username,
                    style = MaterialTheme.typography.titleSmall.copy(fontWeight = FontWeight.Bold),
                    color = MaterialTheme.colorScheme.onSurface
                )
                Text(
                    text = "${entry.matchesPlayed} matches",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }

            // Rating
            Text(
                text = "${entry.rating}",
                style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Black),
                color = MaterialTheme.colorScheme.primary
            )
        }
    }
}
