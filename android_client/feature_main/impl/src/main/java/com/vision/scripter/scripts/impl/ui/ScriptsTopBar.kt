package com.vision.scripter.scripts.impl.ui

import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.vision.scripter.ui.TopBar
import com.vision.scripter.ui.customClickable

@Composable
internal fun ScriptsTopBar(
    node: String,
    onBack: () -> Unit,
) {
    TopBar(
        startContent = {
            Icon(
                modifier = Modifier
                    .size(32.dp)
                    .customClickable(onClick = onBack),
                imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                tint = MaterialTheme.colorScheme.onSurface,
                contentDescription = "",
            )
            Spacer(modifier = Modifier.width(8.dp))
            Text(
                text = node,
                style = TextStyle(
                    fontSize = 22.sp,
                    fontWeight = FontWeight.Bold,
                ),
            )
        },
    )
}

@Preview
@Composable
private fun ScriptsTopBarPreview() {
    ScriptsTopBar(
        node = "Node",
        onBack = {},
    )
}
