package com.vision.scripter.streaming.impl.blocks.video.state

import android.annotation.SuppressLint
import android.media.MediaFormat
import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.Event
import com.vision.scripter.data.api.models.Parameter
import com.vision.scripter.data.api.models.RectangleWithText
import com.vision.scripter.data.api.models.ScreenSizes
import com.vision.scripter.data.api.models.StreamingData
import com.vision.scripter.streaming.impl.screen.main.state.DEFAULT_TIMEOUT
import com.vision.scripter.streaming.impl.screen.main.state.KeyboardState
import com.vision.scripter.streaming.impl.screen.main.state.NEW_SCRIPTS_NODE

data class VideoState(
    val serial: String = "",
    val screenSizes: ScreenSizes? = null,

    val streamingHost: String = "",
    val videoCodec: VideoCodec = VideoCodec.H264,
    val streamingData: StreamingData? = null,
    val cvRectangles: List<CvRectangle> = listOf(),
    val selectedRectangles: List<CvRectangle> = listOf(),

    val record: Record = Record(),
    val keyboard: Keyboard = Keyboard(),
    val tmpParam: Parameter? = null,
) {
    data class Record(
        val controlRecording: Boolean = false,
        val recordName: String = "",
        val node: String = NEW_SCRIPTS_NODE,
        val params: List<Parameter> = listOf(),
        val events: List<Event> = listOf(),
        val timeout: Int = DEFAULT_TIMEOUT,
    )

    data class Keyboard(
        val locale: String = "",
        val buttons: List<RectangleWithText> = listOf(),
        val mode: KeyboardState = KeyboardState.TYPING,
        val typedText: String = "",
    )
}

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
