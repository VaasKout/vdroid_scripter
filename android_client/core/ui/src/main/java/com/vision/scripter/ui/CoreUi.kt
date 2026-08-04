package com.vision.scripter.ui

import androidx.compose.animation.core.LinearEasing
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxScope
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Snackbar
import androidx.compose.material3.SnackbarData
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.material3.pulltorefresh.PullToRefreshDefaults.IndicatorBox
import androidx.compose.material3.pulltorefresh.PullToRefreshState
import androidx.compose.material3.pulltorefresh.rememberPullToRefreshState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.rotate
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

@Composable
fun CustomButton(
    modifier: Modifier = Modifier,
    text: String,
    onClick: () -> Unit,
) {
    Button(
        modifier = modifier,
        shape = RoundedCornerShape(16.dp),
        elevation = ButtonDefaults.elevatedButtonElevation(),
        onClick = onClick,
    ) {
        Text(
            modifier = Modifier.padding(vertical = 4.dp),
            text = text,
            style = TextStyle(
                color = Color.White,
                fontSize = 18.sp,
                fontWeight = FontWeight.Bold
            )
        )
    }
}

@Composable
fun CustomSnackbar(
    snackbarData: SnackbarData,
) {
    Snackbar(
        modifier = Modifier
            .fillMaxWidth()
            .padding(8.dp),
        containerColor = Color.Red,
        shape = RoundedCornerShape(16.dp),
        action = {},
    ) {
        Text(
            modifier = Modifier.padding(8.dp),
            text = snackbarData.visuals.message,
            style = TextStyle(
                color = Color.White,
                fontSize = 16.sp,
                fontWeight = FontWeight.Normal
            )
        )
    }
}

@Composable
fun ProvideSnackbarHost(snackbarHostState: SnackbarHostState) {
    SnackbarHost(
        hostState = snackbarHostState,
        snackbar = {
            CustomSnackbar(snackbarData = it)
        })
}

@Composable
fun CustomPullToRefresh(
    modifier: Modifier = Modifier,
    isRefreshing: Boolean,
    pullToRefreshState: PullToRefreshState = rememberPullToRefreshState(),
    onRefresh: () -> Unit,
    content: @Composable BoxScope.() -> Unit
) {
    val infiniteTransition = rememberInfiniteTransition()
    val rotation by infiniteTransition.animateFloat(
        initialValue = 0f,
        targetValue = 360f,
        animationSpec = infiniteRepeatable(
            animation = tween(300, easing = LinearEasing),
            repeatMode = RepeatMode.Restart
        )
    )
    val rotationByState = if (isRefreshing) rotation
    else pullToRefreshState.distanceFraction * 180f

    PullToRefreshBox(
        modifier = modifier.background(color = MaterialTheme.colorScheme.background),
        isRefreshing = isRefreshing,
        state = pullToRefreshState,
        indicator = {
            IndicatorBox(
                modifier = Modifier.align(Alignment.TopCenter),
                state = pullToRefreshState,
                isRefreshing = isRefreshing,
                containerColor = Color.White,
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(Color.White)
                        .rotate(rotationByState)
                ) {
                    Image(
                        modifier = Modifier.align(Alignment.Center),
                        painter = painterResource(id = R.drawable.cat),
                        contentDescription = "",
                    )
                }
            }
        },
        onRefresh = onRefresh,
        content = content,
    )
}

@Composable
fun DeleteDialog(
    modifier: Modifier = Modifier,
    title: String = "",
    text: String = "",
    onDismiss: () -> Unit,
    onConfirm: () -> Unit,
) {
    AlertDialog(
        modifier = modifier,
        icon = {
            Icon(
                imageVector = Icons.Filled.Delete,
                tint = MaterialTheme.colorScheme.error,
                contentDescription = null,
            )
        },
        title = {
            Text(text = title)
        },
        text = {
            Text(text = text)
        },
        onDismissRequest = onDismiss,
        confirmButton = {
            TextButton(
                onClick = onConfirm,
                colors = ButtonDefaults.textButtonColors(
                    contentColor = MaterialTheme.colorScheme.error,
                ),
            ) {
                Text(text = stringResource(R.string.ok))
            }
        },
        dismissButton = {
            TextButton(
                onClick = onDismiss,
            ) {
                Text(text = stringResource(R.string.cancel))
            }
        }
    )
}

@Composable
fun CommonDialog(
    title: String,
    text: String,
    confirmButtonText: String,
    dismissButtonText: String,
    onDismiss: () -> Unit,
    onConfirm: () -> Unit,
) {
    AlertDialog(
        title = {
            Text(text = title)
        },
        text = {
            Text(text = text)
        },
        onDismissRequest = onDismiss,
        confirmButton = {
            TextButton(onClick = onConfirm) {
                Text(text = confirmButtonText)
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text(text = dismissButtonText)
            }
        },
    )
}

@Composable
fun TopBar(
    modifier: Modifier = Modifier,
    startContent: @Composable () -> Unit = {},
    endContent: @Composable () -> Unit = {},
) {
    Box(
        modifier = modifier
            .fillMaxWidth()
            .background(color = MaterialTheme.colorScheme.primary)
            .statusBarsPadding()
    ) {
        Row(
            modifier = Modifier
                .align(Alignment.CenterStart)
                .padding(16.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            startContent()
        }
        Row(
            modifier = Modifier
                .align(Alignment.CenterEnd)
                .padding(16.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            endContent()
        }
    }
}

@Composable
fun SimpleDropdownMenu(
    options: List<String>,
    onSelect: (String) -> Unit,
) {
    if (options.isEmpty()) return

    var expanded by remember {
        mutableStateOf(false)
    }

    var selectedText by remember {
        mutableStateOf(options.first())
    }

    Box {
        Button(
            onClick = {
                expanded = true
            }
        ) {
            Text(text = selectedText)
        }

        DropdownMenu(
            expanded = expanded,
            onDismissRequest = {
                expanded = false
            }
        ) {
            options.forEach { item ->
                DropdownMenuItem(
                    text = {
                        Text(item)
                    },

                    onClick = {
                        selectedText = item
                        expanded = false
                        onSelect(item)
                    }
                )
            }
        }
    }
}