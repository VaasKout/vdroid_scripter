package com.vision.scripter.streaming.impl.blocks.menu.ui.dialogs

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.text.KeyboardOptions
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
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.vision.scripter.streaming.impl.R
import com.vision.scripter.streaming.impl.screen.main.state.DEFAULT_TIMEOUT
import com.vision.scripter.ui.R as CoreR

private const val MAX_TIMEOUT_DIGITS = 3

@Composable
fun TimeoutDialog(
    initialTimeout: Int,
    onSave: (Int) -> Unit,
    onDismiss: () -> Unit,
) {
    var timeout by remember {
        mutableStateOf(initialTimeout.toString())
    }
    AlertDialog(
        onDismissRequest = onDismiss,
        title = {
            Text(
                modifier = Modifier.padding(8.dp),
                text = stringResource(R.string.insert_timeout),
                style = TextStyle(
                    color = Color.Black,
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Normal
                )
            )
        },
        text = {
            OutlinedTextField(
                value = timeout,
                onValueChange = { input ->
                    timeout = input.filter { it.isDigit() }.take(MAX_TIMEOUT_DIGITS)
                },
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                singleLine = true,
            )
        },
        confirmButton = {
            TextButton(onClick = { onSave(timeout.toIntOrNull() ?: 0) }) {
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
private fun TimeoutDialogPreview() {
    Box(modifier = Modifier.fillMaxSize()) {
        TimeoutDialog(
            initialTimeout = DEFAULT_TIMEOUT,
            onSave = {},
            onDismiss = {},
        )
    }
}
