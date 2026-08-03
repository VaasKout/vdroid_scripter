package com.vision.scripter.editscript.impl.ui

import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.vision.scripter.editscript.impl.R
import com.vision.scripter.ui.TopBar
import com.vision.scripter.ui.customClickable

@Composable
internal fun EditScriptTopBar(
    name: String,
    onBackClick: () -> Unit,
    onSaveClick: () -> Unit,
) {
    TopBar(
        startContent = {
            Icon(
                modifier = Modifier
                    .size(32.dp)
                    .customClickable(onClick = onBackClick),
                imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                tint = MaterialTheme.colorScheme.onSurface,
                contentDescription = "",
            )
            Spacer(modifier = Modifier.width(8.dp))
            Text(
                text = name,
                style = TextStyle(
                    fontSize = 22.sp,
                    fontWeight = FontWeight.Bold,
                ),
            )
        },
        endContent = {
            Text(
                modifier = Modifier
                    .customClickable(onClick = onSaveClick)
                    .padding(horizontal = 8.dp, vertical = 4.dp),
                text = stringResource(R.string.save_button),
                style = TextStyle(
                    fontSize = 18.sp,
                    fontWeight = FontWeight.Bold,
                ),
            )
        },
    )
}

@Preview
@Composable
private fun EditScriptTopBarPreview() {
    EditScriptTopBar(
        name = "login_tap",
        onBackClick = {},
        onSaveClick = {},
    )
}
