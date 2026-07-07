package com.vision.scripter.streaming.impl.blocks.video

import android.annotation.SuppressLint
import android.media.MediaFormat

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