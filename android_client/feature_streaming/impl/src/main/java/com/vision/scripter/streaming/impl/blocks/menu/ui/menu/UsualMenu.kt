package com.vision.scripter.streaming.impl.blocks.menu.ui.menu

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.ArrowForward
import androidx.compose.material.icons.automirrored.filled.Logout
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Camera
import androidx.compose.material.icons.filled.Keyboard
import androidx.compose.material.icons.filled.TextFields
import androidx.compose.material.icons.filled.Visibility
import androidx.compose.material.icons.filled.VisibilityOff
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import com.vision.scripter.streaming.impl.blocks.menu.state.MenuUiStateHolder
import com.vision.scripter.streaming.impl.screen.main.state.CVMode
import com.vision.scripter.streaming.impl.screen.main.state.MenuState
import com.vision.scripter.streaming.impl.screen.main.ui.MenuPreviewUiStateHolder
import com.vision.scripter.ui.customClickable

@Composable
fun UsualMenu(
    modifier: Modifier = Modifier,
    menuState: MenuState.Usual,
    uiStateHolder: MenuUiStateHolder,
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
                        onClick = uiStateHolder::onScriptModeClicked,
                    ),
                imageVector = Icons.Filled.Add,
                tint = MaterialTheme.colorScheme.onSurface,
                contentDescription = ""
            )

            Icon(
                modifier = Modifier
                    .size(32.dp)
                    .customClickable(onClick = uiStateHolder::onCvModeClicked),
                imageVector = detectionIcon(menuState.localCvMode),
                tint = if (menuState.localCvMode != CVMode.NO_CV) Color.Red
                else MaterialTheme.colorScheme.onSurface,
                contentDescription = "",
            )

            Icon(
                modifier = Modifier
                    .size(32.dp)
                    .customClickable(onClick = uiStateHolder::onTextModeClicked),
                imageVector = Icons.Filled.TextFields,
                tint = if (menuState.textHighlighted) Color.Red
                else MaterialTheme.colorScheme.onSurface,
                contentDescription = "",
            )

            Icon(
                modifier = Modifier
                    .size(32.dp)
                    .customClickable(onClick = uiStateHolder::onKeyboardClicked),
                imageVector = Icons.Filled.Keyboard,
                tint = if (menuState.keyboardHighlighted) Color.Red
                else MaterialTheme.colorScheme.onSurface,
                contentDescription = "",
            )

            Icon(
                modifier = Modifier
                    .size(32.dp)
                    .customClickable(onClick = uiStateHolder::onExitClicked),
                imageVector = Icons.AutoMirrored.Filled.Logout,
                tint = MaterialTheme.colorScheme.onSurface,
                contentDescription = ""
            )

            Icon(
                modifier = Modifier
                    .size(32.dp)
                    .customClickable(onClick = uiStateHolder::onExpandClicked),
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
                onClick = uiStateHolder::onExpandClicked,
            ),
        imageVector = Icons.AutoMirrored.Filled.ArrowBack,
        tint = MaterialTheme.colorScheme.onSurface,
        contentDescription = ""
    )
}

fun detectionIcon(cvMode: CVMode): ImageVector = when (cvMode) {
    CVMode.YOLO -> Icons.Filled.Camera
    CVMode.CV_RECTS -> Icons.Filled.Visibility
    else -> Icons.Filled.VisibilityOff
}

@Preview
@Composable
private fun UsualMenuPreview() {
    UsualMenu(
        menuState = MenuState.Usual(),
        uiStateHolder = MenuPreviewUiStateHolder(),
    )
}

@Preview
@Composable
private fun UsualExpandedMenuPreview() {
    UsualMenu(
        menuState = MenuState.Usual(expanded = true),
        uiStateHolder = MenuPreviewUiStateHolder(),
    )
}

@Preview
@Composable
private fun UsualExpandedHighlightedTextPreview() {
    UsualMenu(
        menuState = MenuState.Usual(
            expanded = true,
            localCvMode = CVMode.CV_RECTS,
        ),
        uiStateHolder = MenuPreviewUiStateHolder(),
    )
}
