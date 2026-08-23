package com.examarena.core.designsystem.components

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.examarena.core.designsystem.theme.ElectricCyan
import com.examarena.core.designsystem.theme.PrimaryIndigo
import com.examarena.core.designsystem.theme.VividPurple

@Composable
fun AppAvatar(
    name: String,
    modifier: Modifier = Modifier,
    size: Dp = 48.dp,
    showBorder: Boolean = true,
    rating: Int? = null
) {
    val initials = name.trim().take(2).uppercase()

    val borderBrush = when {
        rating != null && rating >= 2000 -> Brush.linearGradient(listOf(Color(0xFFFFD700), Color(0xFFFFA500)))
        rating != null && rating >= 1600 -> Brush.linearGradient(listOf(VividPurple, ElectricCyan))
        else -> Brush.linearGradient(listOf(PrimaryIndigo, ElectricCyan))
    }

    Box(
        modifier = modifier
            .size(size)
            .then(
                if (showBorder) {
                    Modifier.border(2.dp, borderBrush, CircleShape)
                } else Modifier
            )
            .clip(CircleShape)
            .background(
                Brush.linearGradient(
                    listOf(PrimaryIndigo, VividPurple)
                )
            ),
        contentAlignment = Alignment.Center
    ) {
        Text(
            text = initials.ifEmpty { "EA" },
            color = Color.White,
            fontWeight = FontWeight.Bold,
            fontSize = (size.value * 0.38f).sp
        )
    }
}
