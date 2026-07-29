package com.vision.scripter.streaming.impl.blocks.menu.ui.menu

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Edit
import androidx.compose.material.icons.filled.Keyboard
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import com.vision.scripter.streaming.impl.blocks.menu.state.MenuState
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuPreviewUiStateHolder
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiStateHolder
import com.vision.scripter.streaming.impl.blocks.menu.ui.usualMenuPreviewUiState
import com.vision.scripter.streaming.impl.screen.state.KeyboardState
import com.vision.scripter.ui.customClickable

@Composable
fun KeyboardMenu(
    modifier: Modifier = Modifier,
    menuState: MenuState.Keyboard,
    uiStateHolder: MenuUiStateHolder,
) {
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
        if (menuState.isLoading) {
            CircularProgressIndicator(modifier = Modifier.size(32.dp))
            return
        }

        Icon(
            modifier = Modifier
                .size(32.dp)
                .customClickable(onClick = uiStateHolder::onKeyboardModeClicked),
            imageVector = when (menuState.mode) {
                KeyboardState.TYPING -> Icons.Filled.Keyboard
                KeyboardState.EDIT -> Icons.Filled.Edit
                KeyboardState.ADD_NEW -> Icons.Filled.Add
            },
            tint = Color.Red,
            contentDescription = "",
        )

        Icon(
            modifier = Modifier
                .size(32.dp)
                .customClickable(onClick = uiStateHolder::onSaveClicked),
            imageVector = Icons.Filled.Check,
            tint = Color.Green,
            contentDescription = "",
        )

        Icon(
            modifier = Modifier
                .size(32.dp)
                .customClickable(onClick = uiStateHolder::onCancelClicked),
            imageVector = Icons.AutoMirrored.Filled.ArrowBack,
            tint = MaterialTheme.colorScheme.onSurface,
            contentDescription = "",
        )
    }
}

@Preview
@Composable
private fun KeyboardMenuTypingPreview() {
    KeyboardMenu(
        menuState = MenuState.Keyboard(isLoading = false),
        uiStateHolder = MenuPreviewUiStateHolder(usualMenuPreviewUiState),
    )
}

@Preview
@Composable
private fun KeyboardMenuEditPreview() {
    KeyboardMenu(
        menuState = MenuState.Keyboard(
            isLoading = false,
            mode = KeyboardState.EDIT,
        ),
        uiStateHolder = MenuPreviewUiStateHolder(usualMenuPreviewUiState),
    )
}

@Preview
@Composable
private fun KeyboardMenuAddNewPreview() {
    KeyboardMenu(
        menuState = MenuState.Keyboard(
            isLoading = false,
            mode = KeyboardState.ADD_NEW,
        ),
        uiStateHolder = MenuPreviewUiStateHolder(usualMenuPreviewUiState),
    )
}
