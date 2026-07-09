package com.vision.scripter.scripts.impl.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.vision.scripter.ui.customClickable

@Composable
fun ScriptItem(
    modifier: Modifier = Modifier,
    name: String,
    onPlayClick: (String) -> Unit,
    onDeleteClick: (String) -> Unit,
) {
    Box(
        modifier = modifier
            .fillMaxWidth()
            .background(color = Color.White, shape = RoundedCornerShape(16.dp))
            .padding(16.dp),
    ) {
        Text(
            modifier = Modifier.align(Alignment.Center),
            text = name,
            color = Color.Black,
            style = TextStyle(
                color = Color.Black,
                fontSize = 24.sp,
                fontWeight = FontWeight.Bold
            )
        )

        Row(
            modifier = Modifier.align(Alignment.CenterEnd),
            horizontalArrangement = Arrangement.spacedBy(4.dp),
        ) {
            Icon(
                modifier = Modifier
                    .size(32.dp)
                    .customClickable(onClick = { onDeleteClick(name) }),
                imageVector = Icons.Filled.Delete,
                tint = Color.Red,
                contentDescription = ""
            )
            Icon(
                modifier = Modifier
                    .size(32.dp)
                    .customClickable(
                        onClick = {
                            onPlayClick(name)
                        },
                    ),
                imageVector = Icons.Filled.PlayArrow,
                tint = Color.Green,
                contentDescription = ""
            )
        }
    }
}

@Preview
@Composable
private fun ScriptItemPreview() {
    ScriptItem(
        modifier = Modifier.fillMaxWidth(),
        name = "test_1",
        onPlayClick = {},
        onDeleteClick = {},
    )
}