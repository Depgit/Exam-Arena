package com.examarena.core.designsystem.components

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.examarena.core.designsystem.theme.ErrorRed
import com.examarena.core.designsystem.theme.ErrorRedBg
import com.examarena.core.designsystem.theme.PrimaryIndigo
import com.examarena.core.designsystem.theme.SuccessGreen
import com.examarena.core.designsystem.theme.SuccessGreenBg
import com.examarena.core.designsystem.theme.WarningAmber
import com.examarena.core.designsystem.theme.WarningAmberBg

enum class BadgeVariant {
    Primary,
    Success,
    Warning,
    Error,
    Neutral
}

@Composable
fun AppBadge(
    text: String,
    modifier: Modifier = Modifier,
    variant: BadgeVariant = BadgeVariant.Primary,
    icon: ImageVector? = null
) {
    val (bgColor, textColor, borderColor) = when (variant) {
        BadgeVariant.Primary -> Triple(
            PrimaryIndigo.copy(alpha = 0.12f),
            PrimaryIndigo,
            PrimaryIndigo.copy(alpha = 0.3f)
        )
        BadgeVariant.Success -> Triple(
            SuccessGreenBg,
            SuccessGreen,
            SuccessGreen.copy(alpha = 0.3f)
        )
        BadgeVariant.Warning -> Triple(
            WarningAmberBg,
            WarningAmber,
            WarningAmber.copy(alpha = 0.3f)
        )
        BadgeVariant.Error -> Triple(
            ErrorRedBg,
            ErrorRed,
            ErrorRed.copy(alpha = 0.3f)
        )
        BadgeVariant.Neutral -> Triple(
            MaterialTheme.colorScheme.surfaceVariant,
            MaterialTheme.colorScheme.onSurfaceVariant,
            MaterialTheme.colorScheme.outline.copy(alpha = 0.3f)
        )
    }

    Box(
        modifier = modifier
            .clip(RoundedCornerShape(20.dp))
            .background(bgColor)
            .border(1.dp, borderColor, RoundedCornerShape(20.dp))
            .padding(horizontal = 10.dp, vertical = 4.dp),
        contentAlignment = Alignment.Center
    ) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            if (icon != null) {
                Icon(
                    imageVector = icon,
                    contentDescription = null,
                    modifier = Modifier.size(14.dp),
                    tint = textColor
                )
                Spacer(modifier = Modifier.width(4.dp))
            }
            Text(
                text = text,
                color = textColor,
                fontSize = 12.sp,
                fontWeight = FontWeight.SemiBold
            )
        }
    }
}
