package com.vision.scripter.editscript.impl.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import com.vision.scripter.editscript.impl.R
import com.vision.scripter.ui.customClickable

@Composable
internal fun SectionTitle(
    modifier: Modifier = Modifier,
    text: String,
) {
    Text(
        modifier = modifier.padding(top = 8.dp),
        text = text,
        style = MaterialTheme.typography.titleMedium,
        color = MaterialTheme.colorScheme.onSurfaceVariant,
    )
}

@Composable
internal fun ParamCard(
    modifier: Modifier = Modifier,
    param: ParamUiData,
    onDeleteClick: (Int) -> Unit,
) {
    Box(
        modifier = modifier
            .fillMaxWidth()
            .background(
                color = MaterialTheme.colorScheme.surfaceContainer,
                shape = RoundedCornerShape(16.dp),
            )
            .padding(16.dp),
    ) {
        Column(
            modifier = Modifier
                .align(Alignment.CenterStart)
                .padding(end = 48.dp),
        ) {
            Text(
                text = param.title,
                style = MaterialTheme.typography.titleMedium,
                color = MaterialTheme.colorScheme.onSurface,
            )
            if (param.locale.isNotEmpty()) {
                Text(
                    text = param.locale,
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }
        }

        Icon(
            modifier = Modifier
                .align(Alignment.CenterEnd)
                .size(32.dp)
                .customClickable(onClick = { onDeleteClick(param.id) }),
            imageVector = Icons.Filled.Delete,
            tint = MaterialTheme.colorScheme.error,
            contentDescription = "",
        )
    }
}

@Composable
internal fun DeleteEventsButton(
    modifier: Modifier = Modifier,
    enabled: Boolean,
    onClick: () -> Unit,
) {
    Button(
        modifier = modifier,
        enabled = enabled,
        shape = RoundedCornerShape(16.dp),
        colors = ButtonDefaults.buttonColors(
            containerColor = MaterialTheme.colorScheme.error,
            contentColor = MaterialTheme.colorScheme.onError,
        ),
        onClick = onClick,
    ) {
        Text(text = stringResource(R.string.delete_events_button))
    }
}

@Preview
@Composable
private fun SectionTitlePreview() {
    SectionTitle(text = "Parameters")
}

@Preview
@Composable
private fun ParamCardPreview() {
    ParamCard(
        modifier = Modifier.fillMaxWidth(),
        param = ParamUiData(id = 0, title = "text: Sign in", locale = "eng"),
        onDeleteClick = {},
    )
}

@Preview
@Composable
private fun DeleteEventsButtonPreview() {
    DeleteEventsButton(
        modifier = Modifier.fillMaxWidth(),
        enabled = true,
        onClick = {},
    )
}

@Preview
@Composable
private fun DeleteEventsButtonDisabledPreview() {
    DeleteEventsButton(
        modifier = Modifier.fillMaxWidth(),
        enabled = false,
        onClick = {},
    )
}
