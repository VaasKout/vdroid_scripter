package com.vision.scripter.streaming.impl.ui.items

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.vision.scripter.streaming.impl.R
import com.vision.scripter.streaming.impl.state.keyboardLocales
import com.vision.scripter.ui.SimpleDropdownMenu
import com.vision.scripter.ui.R as CoreR

// TODO move to core
@Composable
fun KeyboardDialog(
    onSaveLocale: (String) -> Unit,
    onDismiss: () -> Unit,
) {
    var selectedLocale by remember { mutableStateOf(keyboardLocales.first()) }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = {
            Text(
                modifier = Modifier.padding(8.dp),
                text = stringResource(R.string.choose_locale),
                style = TextStyle(
                    color = Color.Black,
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Normal
                )
            )
        },
        text = {
            SimpleDropdownMenu(
                options = keyboardLocales,
                onSelect = {
                    selectedLocale = it
                }
            )
        },
        confirmButton = {
            TextButton(onClick = { onSaveLocale(selectedLocale) }) {
                Text(
                    text = stringResource(CoreR.string.ok),
                    style = TextStyle(
                        color = MaterialTheme.colorScheme.secondary,
                        fontSize = 12.sp,
                        fontWeight = FontWeight.Normal
                    )
                )
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text(
                    text = stringResource(CoreR.string.cancel),
                    style = TextStyle(
                        color = MaterialTheme.colorScheme.secondary,
                        fontSize = 12.sp,
                        fontWeight = FontWeight.Normal
                    )
                )
            }
        }
    )
}

@Preview
@Composable
private fun KeyboardDialogPreview() {
    Box(modifier = Modifier.fillMaxSize()) {
        KeyboardDialog(
            onSaveLocale = {},
            onDismiss = {},
        )
    }
}