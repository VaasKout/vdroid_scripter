package com.vision.scripter.streaming.impl.ui.items

import androidx.compose.foundation.Canvas
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.drawscope.DrawScope
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.text.TextMeasurer
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.drawText
import androidx.compose.ui.text.rememberTextMeasurer
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.RectangleWithText

@Composable
fun RectanglesCanvas(
    modifier: Modifier = Modifier,
    cvRectangles: List<CvRectangle>,
    selectedRectangles: List<CvRectangle>,
    keyboardButtons: List<RectangleWithText>,
) {
    val textMeasurer = rememberTextMeasurer()
    Canvas(modifier = modifier) {
        if (keyboardButtons.isNotEmpty()) {
            keyboardButtons.forEach {
                drawRectangle(rectangle = it.rectangle, color = Color.Red)
                drawTextInRectangle(
                    textMeasurer = textMeasurer,
                    rectangle = it.rectangle,
                    text = it.text,
                )
            }
            return@Canvas
        }

        cvRectangles.forEach {
            drawRectangle(rectangle = it, color = Color.Red)
            if (it.label.isNotEmpty()) {
                drawLabel(
                    textMeasurer = textMeasurer,
                    rectangle = it,
                    label = it.label,
                )
            }
        }

        selectedRectangles.forEach {
            drawRectangle(rectangle = it, color = Color.Blue)
        }
    }
}

private fun DrawScope.drawLabel(
    textMeasurer: TextMeasurer,
    rectangle: CvRectangle,
    label: String,
) {
    val measured = textMeasurer.measure(
        text = label,
        style = TextStyle(color = Color.White, fontSize = 10.sp),
    )
    val padding = 2.dp.toPx()
    val left = rectangle.leftX.toFloat()
    val top = rectangle.topY.toFloat()

    drawRect(
        color = Color.Red,
        topLeft = Offset(left, top),
        size = Size(
            width = measured.size.width + padding * 2,
            height = measured.size.height + padding * 2,
        ),
    )
    drawText(
        textLayoutResult = measured,
        topLeft = Offset(left + padding, top + padding),
    )
}

private fun DrawScope.drawTextInRectangle(
    textMeasurer: TextMeasurer,
    rectangle: CvRectangle?,
    text: String,
) {
    rectangle ?: return

    val x = (rectangle.leftX + 12).toFloat()
    val y = (rectangle.topY + 12).toFloat()

    drawText(
        textMeasurer = textMeasurer,
        text = text,
        topLeft = Offset(x, y),
        style = TextStyle(
            color = Color.Red,
            fontSize = 12.sp,
        )
    )
}

private fun DrawScope.drawRectangle(
    rectangle: CvRectangle?,
    color: Color,
) {
    rectangle ?: return
    val width = (rectangle.rightX - rectangle.leftX).toFloat()
    val height = (rectangle.bottomY - rectangle.topY).toFloat()
    drawRect(
        color = color,
        topLeft = Offset(rectangle.leftX.toFloat(), rectangle.topY.toFloat()),
        size = Size(width = width, height = height),
        style = Stroke(
            width = 1.dp.toPx()
        )
    )
}