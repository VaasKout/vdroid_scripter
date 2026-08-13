package com.vision.scripter.main.ui.items

import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.tooling.preview.Preview
import com.vision.scripter.ui.TopBar
import com.vision.scripter.ui.customClickable

@Composable
internal fun MainTopBar(
    onSettingsClick: () -> Unit
) {
    TopBar(
        startContent = {},
        endContent = {
            Icon(
                modifier = Modifier.customClickable(
                    onClick = onSettingsClick,
                ),
                imageVector = Icons.Default.Settings,
                tint = MaterialTheme.colorScheme.onSurface,
                contentDescription = ""
            )
        }
    )
}

@Preview
@Composable
private fun MainTopBarPreview() {
    MainTopBar(
        onSettingsClick = {}
    )
}