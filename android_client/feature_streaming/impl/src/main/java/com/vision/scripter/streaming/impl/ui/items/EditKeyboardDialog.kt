package com.vision.scripter.streaming.impl.ui.items

import androidx.compose.foundation.layout.Box
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
import com.vision.scripter.ui.R as CoreR

private const val MAX_BUTTON_NAME_LENGTH = 32

@Composable
fun EditKeyboardDialog(
    onSave: (String) -> Unit,
    onDismiss: () -> Unit,
) {
    var text by remember { mutableStateOf("") }
    var isError by remember { mutableStateOf(false) }
    AlertDialog(
        onDismissRequest = onDismiss,
        title = {
            Text(
                modifier = Modifier.padding(8.dp),
                text = stringResource(R.string.set_button),
                style = TextStyle(
                    color = Color.Black,
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Normal,
                ),
            )
        },
        text = {
            OutlinedTextField(
                value = text,
                onValueChange = { new ->
                    if (new.length <= MAX_BUTTON_NAME_LENGTH) {
                        text = new
                        if (new.isNotBlank()) isError = false
                    }
                },
                label = {
                    Text(text = stringResource(CoreR.string.name))
                },
                isError = isError,
                singleLine = true,
            )
        },
        confirmButton = {
            TextButton(
                onClick = {
                    if (text.isBlank()) {
                        isError = true
                    } else {
                        onSave(text)
                    }
                },
            ) {
                Text(
                    text = stringResource(CoreR.string.ok),
                    style = TextStyle(
                        color = MaterialTheme.colorScheme.secondary,
                        fontSize = 12.sp,
                        fontWeight = FontWeight.Normal,
                    ),
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
                        fontWeight = FontWeight.Normal,
                    ),
                )
            }
        },
    )
}

@Preview
@Composable
private fun EditKeyboardDialogPreview() {
    Box(modifier = Modifier.fillMaxSize()) {
        EditKeyboardDialog(
            onSave = {},
            onDismiss = {},
        )
    }
}
