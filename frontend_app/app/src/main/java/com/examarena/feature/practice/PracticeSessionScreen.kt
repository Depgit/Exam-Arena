package com.examarena.feature.practice

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.background
import androidx.compose.foundation.border
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
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.EmojiEvents
import androidx.compose.material.icons.filled.Lightbulb
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
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.examarena.core.designsystem.components.AppButton
import com.examarena.core.designsystem.components.AppCard
import com.examarena.core.designsystem.components.AppTopBar
import com.examarena.core.designsystem.components.CircularAccuracyGauge
import com.examarena.core.designsystem.theme.ErrorRed
import com.examarena.core.designsystem.theme.ErrorRedBg
import com.examarena.core.designsystem.theme.GoldWinner
import com.examarena.core.designsystem.theme.PrimaryIndigo
import com.examarena.core.designsystem.theme.SuccessGreen
import com.examarena.core.designsystem.theme.SuccessGreenBg
import com.examarena.data.model.OptionForPlayer

@Composable
fun PracticeSessionScreen(
    viewModel: PracticeViewModel,
    onBackClick: () -> Unit
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val scrollState = rememberScrollState()

    if (uiState.isSessionCompleted) {
        PracticeCompletionView(
            correctCount = uiState.correctAnswersCount,
            totalCount = uiState.questions.size,
            onDone = onBackClick
        )
        return
    }

    val currentQuestion = uiState.questions.getOrNull(uiState.currentQuestionIndex)
    val totalCount = uiState.questions.size

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        AppTopBar(
            title = "Practice (${uiState.selectedCategory.code})",
            onBackClick = onBackClick
        )

        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(scrollState)
                .padding(horizontal = 20.dp, vertical = 12.dp)
        ) {
            // Question Counter
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Surface(
                    shape = RoundedCornerShape(12.dp),
                    color = PrimaryIndigo.copy(alpha = 0.12f)
                ) {
                    Text(
                        text = "Question ${uiState.currentQuestionIndex + 1} of $totalCount",
                        style = MaterialTheme.typography.labelLarge.copy(fontWeight = FontWeight.Bold),
                        color = PrimaryIndigo,
                        modifier = Modifier.padding(horizontal = 12.dp, vertical = 6.dp)
                    )
                }

                Text(
                    text = "Score: ${uiState.correctAnswersCount} / $totalCount",
                    style = MaterialTheme.typography.labelLarge.copy(fontWeight = FontWeight.Bold),
                    color = SuccessGreen
                )
            }

            Spacer(modifier = Modifier.height(16.dp))

            if (currentQuestion != null) {
                // Question Body Card
                AppCard(
                    modifier = Modifier.fillMaxWidth(),
                    containerColor = MaterialTheme.colorScheme.surface,
                    elevation = 2.dp,
                    contentPadding = 20.dp
                ) {
                    Text(
                        text = currentQuestion.body,
                        style = MaterialTheme.typography.titleMedium.copy(
                            fontWeight = FontWeight.SemiBold,
                            lineHeight = 26.sp
                        ),
                        color = MaterialTheme.colorScheme.onSurface
                    )
                }

                Spacer(modifier = Modifier.height(20.dp))

                // Options List
                val letters = listOf("A", "B", "C", "D", "E")
                currentQuestion.options.forEachIndexed { index, option ->
                    val letter = letters.getOrElse(index) { "${index + 1}" }
                    val isSelected = uiState.selectedOptionId == option.id
                    val isCorrectOption = uiState.evaluationResult?.correctOptionId == option.id

                    PracticeOptionItem(
                        letter = letter,
                        option = option,
                        isSelected = isSelected,
                        isEvaluated = uiState.isAnswerEvaluated,
                        isCorrect = isCorrectOption,
                        onClick = {
                            if (!uiState.isAnswerEvaluated && !uiState.isLoading) {
                                viewModel.submitAnswer(option.id)
                            }
                        }
                    )

                    Spacer(modifier = Modifier.height(12.dp))
                }

                // ── Instant Explanation Card ────────────────────────
                AnimatedVisibility(
                    visible = uiState.isAnswerEvaluated,
                    enter = fadeIn(),
                    exit = fadeOut()
                ) {
                    val eval = uiState.evaluationResult
                    Column(modifier = Modifier.padding(top = 12.dp)) {
                        AppCard(
                            modifier = Modifier.fillMaxWidth(),
                            containerColor = if (eval?.isCorrect == true) SuccessGreenBg else ErrorRedBg,
                            borderColor = if (eval?.isCorrect == true) SuccessGreen else ErrorRed,
                            elevation = 1.dp
                        ) {
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                Icon(
                                    imageVector = if (eval?.isCorrect == true) Icons.Default.CheckCircle else Icons.Default.Close,
                                    contentDescription = null,
                                    tint = if (eval?.isCorrect == true) SuccessGreen else ErrorRed,
                                    modifier = Modifier.size(22.dp)
                                )
                                Spacer(modifier = Modifier.width(8.dp))
                                Text(
                                    text = if (eval?.isCorrect == true) "CORRECT ANSWER! ✓" else "INCORRECT ✕",
                                    style = MaterialTheme.typography.titleSmall.copy(fontWeight = FontWeight.Black),
                                    color = if (eval?.isCorrect == true) SuccessGreen else ErrorRed
                                )
                            }

                            if (!eval?.explanation.isNullOrBlank()) {
                                Spacer(modifier = Modifier.height(10.dp))
                                Row(verticalAlignment = Alignment.Top) {
                                    Icon(
                                        imageVector = Icons.Default.Lightbulb,
                                        contentDescription = null,
                                        tint = PrimaryIndigo,
                                        modifier = Modifier.size(18.dp)
                                    )
                                    Spacer(modifier = Modifier.width(6.dp))
                                    Text(
                                        text = eval?.explanation ?: "",
                                        style = MaterialTheme.typography.bodyMedium,
                                        color = MaterialTheme.colorScheme.onSurface
                                    )
                                }
                            }
                        }

                        Spacer(modifier = Modifier.height(20.dp))

                        AppButton(
                            text = if (uiState.currentQuestionIndex == totalCount - 1) "Complete Practice Session" else "Next Question →",
                            onClick = { viewModel.nextQuestion() }
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.height(40.dp))
        }
    }
}

@Composable
private fun PracticeOptionItem(
    letter: String,
    option: OptionForPlayer,
    isSelected: Boolean,
    isEvaluated: Boolean,
    isCorrect: Boolean,
    onClick: () -> Unit
) {
    val borderColor = when {
        isEvaluated && isCorrect -> SuccessGreen
        isEvaluated && isSelected && !isCorrect -> ErrorRed
        isSelected -> PrimaryIndigo
        else -> MaterialTheme.colorScheme.outline.copy(alpha = 0.5f)
    }

    val bgColor = when {
        isEvaluated && isCorrect -> SuccessGreenBg
        isEvaluated && isSelected && !isCorrect -> ErrorRedBg
        isSelected -> PrimaryIndigo.copy(alpha = 0.12f)
        else -> MaterialTheme.colorScheme.surface
    }

    Surface(
        modifier = Modifier
            .fillMaxWidth()
            .clip(MaterialTheme.shapes.medium)
            .border(
                width = if (isSelected || (isEvaluated && isCorrect)) 2.dp else 1.dp,
                color = borderColor,
                shape = MaterialTheme.shapes.medium
            )
            .clickable(enabled = !isEvaluated, onClick = onClick),
        shape = MaterialTheme.shapes.medium,
        color = bgColor,
        tonalElevation = 1.dp
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 16.dp, vertical = 14.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Box(
                modifier = Modifier
                    .size(32.dp)
                    .clip(CircleShape)
                    .background(
                        when {
                            isEvaluated && isCorrect -> SuccessGreen
                            isEvaluated && isSelected && !isCorrect -> ErrorRed
                            isSelected -> PrimaryIndigo
                            else -> MaterialTheme.colorScheme.surfaceVariant
                        }
                    ),
                contentAlignment = Alignment.Center
            ) {
                if (isEvaluated && isCorrect) {
                    Icon(
                        imageVector = Icons.Default.Check,
                        contentDescription = null,
                        tint = Color.White,
                        modifier = Modifier.size(18.dp)
                    )
                } else if (isEvaluated && isSelected && !isCorrect) {
                    Icon(
                        imageVector = Icons.Default.Close,
                        contentDescription = null,
                        tint = Color.White,
                        modifier = Modifier.size(18.dp)
                    )
                } else {
                    Text(
                        text = letter,
                        style = MaterialTheme.typography.labelLarge.copy(fontWeight = FontWeight.Bold),
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }

            Spacer(modifier = Modifier.width(14.dp))

            Text(
                text = option.optionText,
                style = MaterialTheme.typography.bodyLarge.copy(
                    fontWeight = if (isSelected || (isEvaluated && isCorrect)) FontWeight.Bold else FontWeight.Normal
                ),
                color = MaterialTheme.colorScheme.onSurface,
                modifier = Modifier.weight(1f)
            )
        }
    }
}

@Composable
private fun PracticeCompletionView(
    correctCount: Int,
    totalCount: Int,
    onDone: () -> Unit
) {
    val accuracy = ((correctCount.toFloat() / totalCount.toFloat().coerceAtLeast(1f)) * 100).toInt()

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
            .padding(24.dp),
        contentAlignment = Alignment.Center
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Icon(
                imageVector = Icons.Default.EmojiEvents,
                contentDescription = null,
                tint = GoldWinner,
                modifier = Modifier.size(72.dp)
            )

            Spacer(modifier = Modifier.height(16.dp))

            Text(
                text = "Practice Complete!",
                style = MaterialTheme.typography.headlineLarge.copy(fontWeight = FontWeight.Black),
                color = MaterialTheme.colorScheme.onBackground
            )

            Spacer(modifier = Modifier.height(24.dp))

            CircularAccuracyGauge(
                accuracy = accuracy.toFloat(),
                size = 110.dp,
                strokeWidth = 10.dp
            )

            Spacer(modifier = Modifier.height(20.dp))

            Text(
                text = "You scored $correctCount out of $totalCount questions correctly.",
                style = MaterialTheme.typography.bodyLarge,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )

            Spacer(modifier = Modifier.height(36.dp))

            AppButton(
                text = "Return to Arena Home",
                onClick = onDone
            )
        }
    }
}
