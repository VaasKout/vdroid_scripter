package com.vision.scripter.streaming.impl.blocks.menu.ui.dialogs

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.selection.toggleable
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Checkbox
import androidx.compose.material3.MaterialTheme
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
import com.vision.scripter.streaming.impl.screen.state.locales
import com.vision.scripter.ui.SimpleDropdownMenu
import com.vision.scripter.ui.R as CoreR

@Composable
fun ScanDialog(
    onScan: (locale: String, includeImages: Boolean) -> Unit,
    onDismiss: () -> Unit,
) {
    var selectedLocale by remember { mutableStateOf(locales.first()) }
    var includeImages by remember { mutableStateOf(false) }
    AlertDialog(
        onDismissRequest = onDismiss,
        title = {
            Text(
                modifier = Modifier.padding(8.dp),
                text = stringResource(R.string.scan_screen),
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
                SimpleDropdownMenu(
                    options = locales,
                    onSelect = {
                        selectedLocale = it
                    }
                )
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .toggleable(
                            value = includeImages,
                            onValueChange = { includeImages = it },
                        )
                        .padding(vertical = 4.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Checkbox(
                        checked = includeImages,
                        onCheckedChange = null,
                    )
                    Text(
                        text = stringResource(R.string.include_library_images),
                        style = TextStyle(
                            color = Color.Black,
                            fontSize = 14.sp,
                            fontWeight = FontWeight.Normal
                        )
                    )
                }
            }
        },
        confirmButton = {
            TextButton(onClick = { onScan(selectedLocale, includeImages) }) {
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
private fun ScanDialogPreview() {
    Box(modifier = Modifier.fillMaxSize()) {
        ScanDialog(
            onScan = { _, _ -> },
            onDismiss = {},
        )
    }
}
