package com.vision.scripter.streaming.impl.blocks.menu.ui.dialogs

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
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
import com.vision.scripter.streaming.impl.screen.state.locales
import com.vision.scripter.ui.SimpleDropdownMenu
import com.vision.scripter.ui.R as CoreR

// TODO move to core
@Composable
fun TextToFindDialog(
    onTryToFindText: (String, String) -> Unit,
    onDismiss: () -> Unit,
) {
    var text by remember { mutableStateOf("") }
    var selectedLocale by remember { mutableStateOf(locales.first()) }
    AlertDialog(
        onDismissRequest = onDismiss,
        title = {
            Text(
                modifier = Modifier.padding(8.dp),
                text = stringResource(R.string.write_the_word),
                style = TextStyle(
                    color = Color.Black,
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Normal
                )
            )
        },
        text = {
            Column(
                verticalArrangement = Arrangement.spacedBy(4.dp),
            ) {
                OutlinedTextField(
                    value = text,
                    onValueChange = { text = it },
                    label = {
                        Text(
                            text = stringResource(CoreR.string.name)
                        )
                    },
                    singleLine = true,
                )
                SimpleDropdownMenu(
                    options = locales,
                    onSelect = {
                        selectedLocale = it
                    }
                )
            }
        },
        confirmButton = {
            TextButton(onClick = { onTryToFindText(text, selectedLocale) }) {
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
private fun TextToFindDialogPreview() {
    Box(modifier = Modifier.fillMaxSize()) {
        TextToFindDialog(
            onTryToFindText = { _, _ -> },
            onDismiss = {},
        )
    }
}