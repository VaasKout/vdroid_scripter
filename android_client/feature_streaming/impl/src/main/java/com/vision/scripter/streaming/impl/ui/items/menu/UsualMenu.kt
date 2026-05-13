package com.vision.scripter.streaming.impl.ui.items.menu

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.ArrowForward
import androidx.compose.material.icons.filled.Description
import androidx.compose.material.icons.filled.Keyboard
import androidx.compose.material.icons.filled.VisibilityOff
import androidx.compose.material.icons.rounded.Rectangle
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import com.vision.scripter.streaming.impl.state.CVMode
import com.vision.scripter.streaming.impl.state.MenuState
import com.vision.scripter.ui.customClickable

@Composable
fun UsualMenu(
    modifier: Modifier = Modifier,
    menuState: MenuState.Usual,
    onScriptModeClick: () -> Unit,
    onKeyboardClick: () -> Unit,
    onCvModeClick: () -> Unit,
    onExpandClick: () -> Unit,
) {
    if (menuState.expanded) {
        Column(
            modifier = modifier
                .background(
                    color = Color.White,
                    shape = RoundedCornerShape(topStart = 8.dp, bottomStart = 8.dp),
                )
                .padding(horizontal = 4.dp),
            horizontalAlignment = Alignment.End,
            verticalArrangement = Arrangement.spacedBy(4.dp),
        ) {
            Icon(
                modifier = Modifier
                    .size(32.dp)
                    .customClickable(
                        onClick = onScriptModeClick,
                    ),
                imageVector = Icons.Filled.Description,
                tint = MaterialTheme.colorScheme.onSurface,
                contentDescription = ""
            )

            if (menuState.keyboardLoading) {
                CircularProgressIndicator()
            } else {
                Icon(
                    modifier = Modifier
                        .size(32.dp)
                        .customClickable(onClick = onKeyboardClick),
                    imageVector = Icons.Filled.Keyboard,
                    tint = if (menuState.showKeyboardButtons) Color.Red
                    else MaterialTheme.colorScheme.onSurface,
                    contentDescription = ""
                )
            }

            Icon(
                modifier = Modifier
                    .size(32.dp)
                    .customClickable(
                        onClick = onCvModeClick,
                    ),
                imageVector = when (menuState.cvMode) {
                    CVMode.NO_CV -> Icons.Filled.VisibilityOff
                    CVMode.CV_RECTS -> Icons.Rounded.Rectangle
                },
                tint =
                    if (menuState.cvMode == CVMode.NO_CV) MaterialTheme.colorScheme.onSurface
                    else Color.Red,
                contentDescription = ""
            )

            Icon(
                modifier = Modifier
                    .size(32.dp)
                    .customClickable(onClick = onExpandClick),
                imageVector = Icons.AutoMirrored.Filled.ArrowForward,
                tint = MaterialTheme.colorScheme.onSurface,
                contentDescription = ""
            )
        }

        return
    }

    Icon(
        modifier = modifier
            .size(32.dp)
            .background(
                color = Color.White,
                shape = RoundedCornerShape(topStart = 8.dp, bottomStart = 8.dp),
            )
            .customClickable(
                onClick = onExpandClick,
            ),
        imageVector = Icons.AutoMirrored.Filled.ArrowBack,
        tint = MaterialTheme.colorScheme.onSurface,
        contentDescription = ""
    )
}


@Preview
@Composable
private fun UsualMenuPreview() {
    UsualMenu(
        menuState = MenuState.Usual(),
        onScriptModeClick = {},
        onKeyboardClick = {},
        onCvModeClick = {},
        onExpandClick = {},
    )
}