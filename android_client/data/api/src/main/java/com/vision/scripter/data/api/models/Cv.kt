package com.vision.scripter.data.api.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class SaveRectRequest(
    @SerialName("serial")
    val serial: String,
    @SerialName("node")
    val node: String,
    @SerialName("name")
    val name: String,
    @SerialName("value")
    val value: String,
    @SerialName("rectangle")
    val rectangle: CvRectangle?,
)

@Serializable
data class EditKeyboardRequest(
    @SerialName("serial")
    val serial: String,
    @SerialName("locale")
    val locale: String,
    @SerialName("name")
    val name: String,
    @SerialName("rectangle")
    val rectangle: CvRectangle?,
)

@Serializable
data class CvRectangle(
    @SerialName("left_x")
    val leftX: Int,
    @SerialName("right_x")
    val rightX: Int,
    @SerialName("top_y")
    val topY: Int,
    @SerialName("bottom_y")
    val bottomY: Int,
    @SerialName("label")
    val label: String = "",
)

fun List<CvRectangle>.adjustToClient(
    screenSizes: ScreenSizes,
): List<CvRectangle> = map {
    it.adjustToClient(screenSizes)
}

fun CvRectangle.adjustToClient(
    screenSizes: ScreenSizes,
): CvRectangle {
    val scaleWidth = screenSizes.surfaceWidth.toFloat() / screenSizes.remoteWidth.toFloat()
    val scaleHeight = screenSizes.surfaceHeight.toFloat() / screenSizes.remoteHeight.toFloat()

    val scaledVideoWidth = screenSizes.remoteWidth * scaleWidth
    val scaledVideoHeight = screenSizes.remoteHeight * scaleHeight

    val offsetX = (screenSizes.surfaceWidth - scaledVideoWidth) / 2f
    val offsetY = (screenSizes.surfaceHeight - scaledVideoHeight) / 2f

    return copy(
        leftX = (leftX * scaleWidth + offsetX).toInt(),
        topY = (topY * scaleHeight + offsetY).toInt(),
        rightX = (rightX * scaleWidth + offsetX).toInt(),
        bottomY = (bottomY * scaleHeight + offsetY).toInt(),
    )
}

fun CvRectangle.adjustToServer(
    screenSizes: ScreenSizes,
): CvRectangle {
    val scaleWidth = screenSizes.remoteWidth.toFloat() / screenSizes.surfaceWidth.toFloat()
    val scaleHeight = screenSizes.remoteHeight.toFloat() / screenSizes.surfaceHeight.toFloat()

    val remoteLeftX = leftX * scaleWidth
    val remoteTopY = topY * scaleHeight
    val remoteRightX = rightX * scaleWidth
    val remoteBottomY = bottomY * scaleHeight

    return copy(
        leftX = remoteLeftX.toInt(),
        topY = remoteTopY.toInt(),
        rightX = remoteRightX.toInt(),
        bottomY = remoteBottomY.toInt(),
    )
}

fun CvRectangle?.isEmpty(): Boolean =
    this == null || (leftX == 0 && topY == 0 && rightX == 0 && bottomY == 0)

fun CvRectangle?.contains(x: Int, y: Int): Boolean {
    this ?: return false
    if (this.isEmpty() || (x == 0 && y == 0)) return false
    return x in (leftX..rightX) && y in (topY..bottomY)
}

fun List<CvRectangle>.smallestBy(x: Int, y: Int): CvRectangle? {
    return filter { it.contains(x, y) }.minByOrNull {
        (it.rightX - it.leftX) * (it.bottomY - it.topY)
    }
}