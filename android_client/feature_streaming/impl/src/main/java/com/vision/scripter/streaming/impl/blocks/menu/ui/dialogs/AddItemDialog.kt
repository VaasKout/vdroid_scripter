package com.vision.scripter.streaming.impl.blocks.menu.ui.dialogs

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.selection.selectable
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.RadioButton
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
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

@Composable
fun AddItemDialog(
    onConfirm: (name: String, isImage: Boolean) -> Unit,
    onDismiss: () -> Unit,
) {
    var text by remember { mutableStateOf("") }
    var isImage by remember { mutableStateOf(true) }
    AlertDialog(
        onDismissRequest = onDismiss,
        title = {
            Text(
                modifier = Modifier.padding(8.dp),
                text = stringResource(R.string.add_to_library),
                style = TextStyle(
                    color = Color.Black,
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Normal
                )
            )
        },
        text = {
            Column {
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

                AddItemChoice(
                    text = stringResource(R.string.add_image),
                    selected = isImage,
                    onClick = { isImage = true },
                )
                AddItemChoice(
                    text = stringResource(R.string.add_custom_action),
                    selected = !isImage,
                    onClick = { isImage = false },
                )
            }
        },
        confirmButton = {
            TextButton(onClick = { onConfirm(text, isImage) }) {
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

@Composable
private fun AddItemChoice(
    text: String,
    selected: Boolean,
    onClick: () -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .selectable(
                selected = selected,
                onClick = onClick,
            )
            .padding(vertical = 4.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        RadioButton(
            selected = selected,
            onClick = onClick,
        )
        Text(
            text = text,
            style = TextStyle(
                color = Color.Black,
                fontSize = 14.sp,
                fontWeight = FontWeight.Normal
            )
        )
    }
}

@Preview
@Composable
private fun AddItemDialogPreview() {
    Box(modifier = Modifier.fillMaxSize()) {
        AddItemDialog(
            onConfirm = { _, _ -> },
            onDismiss = {},
        )
    }
}
