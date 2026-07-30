package com.vision.scripter.streaming.impl.blocks.video.state

import android.annotation.SuppressLint
import android.media.MediaFormat
import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.RectangleWithText
import com.vision.scripter.data.api.models.ScreenSizes
import com.vision.scripter.data.api.models.StreamingData
import com.vision.scripter.streaming.impl.data.Record

data class VideoState(
    val serial: String = "",
    val screenSizes: ScreenSizes? = null,

    val streamingHost: String = "",
    val videoCodec: VideoCodec = VideoCodec.H264,
    val streamingData: StreamingData? = null,

    val cvRectangles: List<CvRectangle> = listOf(),
    val selectedRectangles: List<CvRectangle> = listOf(),
    val keyboardButtons: List<RectangleWithText> = listOf(),
    val record: Record = Record(),
)

enum class VideoCodec(
    val id: Int,
    val codecName: String, // 4-byte ASCII representation of the name
    val mimeType: String,
) {
    H264(0x68323634, "h264", MediaFormat.MIMETYPE_VIDEO_AVC),
    H265(0x68323635, "h265", MediaFormat.MIMETYPE_VIDEO_HEVC),

    @SuppressLint("InlinedApi")  // introduced in API 29
    AV1(0x00617631, "av1", MediaFormat.MIMETYPE_VIDEO_AV1);

    companion object {
        fun findByName(name: String): VideoCodec? {
            return entries.firstOrNull { it.codecName == name }
        }
    }
}
