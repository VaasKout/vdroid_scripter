package com.vision.scripter.streaming.impl.blocks.menu.ui.menu

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Edit
import androidx.compose.material.icons.filled.RadioButtonChecked
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import com.vision.scripter.streaming.impl.blocks.menu.state.MenuUiStateHolder
import com.vision.scripter.streaming.impl.screen.main.state.MenuState
import com.vision.scripter.streaming.impl.screen.main.ui.MenuPreviewUiStateHolder
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
        if (menuState.isLoadingKeyboard) {
            CircularProgressIndicator(modifier = Modifier.size(32.dp))
            return
        }

        if (!menuState.fromUsual) {
            Icon(
                modifier = Modifier
                    .size(32.dp)
                    .customClickable(
                        onClick = uiStateHolder::onRecordingClicked,
                    ),
                imageVector = Icons.Filled.RadioButtonChecked,
                tint =
                    if (menuState.recordingKeyboard) Color.Red
                    else MaterialTheme.colorScheme.onSurface,
                contentDescription = "",
            )
        }

        Icon(
            modifier = Modifier
                .size(32.dp)
                .customClickable(onClick = uiStateHolder::onKeyboardInitClicked),
            imageVector = Icons.Filled.Search,
            tint = MaterialTheme.colorScheme.onSurface,
            contentDescription = ""
        )

        Icon(
            modifier = Modifier
                .size(32.dp)
                .customClickable(onClick = { uiStateHolder.onKeyboardEdited(false) }),
            imageVector = Icons.Filled.Edit,
            tint = if (menuState.editing) Color.Red
            else MaterialTheme.colorScheme.onSurface,
            contentDescription = ""
        )

        Icon(
            modifier = Modifier
                .size(32.dp)
                .customClickable(onClick = { uiStateHolder.onKeyboardEdited(true) }),
            imageVector = Icons.Filled.Add,
            tint = if (menuState.showCvRectangles) Color.Red
            else MaterialTheme.colorScheme.onSurface,
            contentDescription = "",
        )

        Icon(
            modifier = Modifier
                .size(32.dp)
                .customClickable(onClick = uiStateHolder::onCancelClicked),
            imageVector = Icons.Filled.Close,
            tint = MaterialTheme.colorScheme.onSurface,
            contentDescription = ""
        )

        if (!menuState.fromUsual) {
            Icon(
                modifier = Modifier
                    .size(32.dp)
                    .customClickable(onClick = uiStateHolder::onSaveClicked),
                imageVector = Icons.Filled.Check,
                tint = Color.Green,
                contentDescription = "",
            )
        }
    }
}

@Preview
@Composable
private fun KeyboardMenuDefaultPreview() {
    KeyboardMenu(
        menuState = MenuState.Keyboard(isLoadingKeyboard = false),
        uiStateHolder = MenuPreviewUiStateHolder(),
    )
}

@Preview
@Composable
private fun KeyboardMenuFromUsualPreview() {
    KeyboardMenu(
        menuState = MenuState.Keyboard(
            isLoadingKeyboard = false,
            fromUsual = true,
        ),
        uiStateHolder = MenuPreviewUiStateHolder(),
    )
}

@Preview
@Composable
private fun KeyboardMenuEditingPreview() {
    KeyboardMenu(
        menuState = MenuState.Keyboard(
            isLoadingKeyboard = false,
            editing = true,
        ),
        uiStateHolder = MenuPreviewUiStateHolder(),
    )
}

@Preview
@Composable
private fun KeyboardMenuEditingShowCvPreview() {
    KeyboardMenu(
        menuState = MenuState.Keyboard(
            isLoadingKeyboard = false,
            editing = true,
            showCvRectangles = true,
        ),
        uiStateHolder = MenuPreviewUiStateHolder(),
    )
}
