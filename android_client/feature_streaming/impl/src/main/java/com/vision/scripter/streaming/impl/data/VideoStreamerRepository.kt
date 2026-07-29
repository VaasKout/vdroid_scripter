package com.vision.scripter.streaming.impl.data

import android.media.MediaCodec
import android.media.MediaCodecInfo
import android.media.MediaFormat
import android.view.Surface
import com.vision.scripter.data.api.VideoStreamer
import com.vision.scripter.data.api.models.ScreenSizes
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import javax.inject.Inject

@ViewModelScoped
class VideoStreamerRepository @Inject constructor(
    private val videoStreamer: VideoStreamer,
) {

    @Volatile
    private var mediaCodec: MediaCodec? = null

    @Volatile
    private var surface: Surface? = null

    private val screenSizesFlow = MutableStateFlow<ScreenSizes?>(null)
    fun observeScreenSizes(): StateFlow<ScreenSizes?> = screenSizesFlow.asStateFlow()

    fun initConnection(
        host: String,
        port: Int,
        newSurface: Surface,
        surfaceWidth: Int,
        surfaceHeight: Int,
    ): Boolean {
        try {
            stop()
            surface = newSurface
            val connected = videoStreamer.connect(host, port)
            if (!connected) return false
            val pair = videoStreamer.readVideoHeader()
            screenSizesFlow.update {
                ScreenSizes(
                    remoteWidth = pair.first,
                    remoteHeight = pair.second,
                    surfaceWidth = surfaceWidth,
                    surfaceHeight = surfaceHeight,
                )
            }
            return true
        } catch (e: Exception) {
            e.printStackTrace()
        }
        return false
    }

    fun decodeFramesInLoop(mimeType: String) {
        mediaCodec = try {
            MediaCodec.createDecoderByType(mimeType)
        } catch (e: Exception) {
            e.printStackTrace()
            null
        } ?: return

        val screenSizes = screenSizesFlow.value ?: return
        val format = createFormat(mimeType).apply {
            setInteger(MediaFormat.KEY_WIDTH, screenSizes.remoteWidth)
            setInteger(MediaFormat.KEY_HEIGHT, screenSizes.remoteHeight)
        }

        try {
            mediaCodec?.apply {
                configure(format, surface, null, 0)
                start()
            }
        } catch (e: Exception) {
            e.printStackTrace()
            return
        }

        while (true) {
            try {
                val (pts, size) = videoStreamer.readFrameMeta()
                val byteArray = videoStreamer.readFrame(size) ?: break

                decodeLoop(byteArray = byteArray, pts = pts)
            } catch (e: Exception) {
                e.printStackTrace()
                break
            }
        }
    }

    private fun decodeLoop(
        byteArray: ByteArray,
        pts: Long,
    ) {
        val bufferInfo = MediaCodec.BufferInfo()

        var extractedPts = pts

        val isConfig = (pts and PACKET_FLAG_CONFIG) != 0L
        if (isConfig) {
            extractedPts = extractedPts and PACKET_FLAG_CONFIG.inv()
        }
        val isKeyFrame = (pts and PACKET_FLAG_KEY_FRAME) != 0L
        if (isKeyFrame) {
            extractedPts = extractedPts and PACKET_FLAG_KEY_FRAME.inv()
        }
        val flag = when {
            isConfig -> MediaCodec.BUFFER_FLAG_CODEC_CONFIG
            isKeyFrame -> MediaCodec.BUFFER_FLAG_KEY_FRAME
            else -> 0
        }

        val inputBufferId = mediaCodec?.dequeueInputBuffer(-1) ?: 0
        if (inputBufferId >= 0) {
            val inputBuffer = mediaCodec?.getInputBuffer(inputBufferId) ?: return
            inputBuffer.clear()

            inputBuffer.put(byteArray, 0, byteArray.size)
            mediaCodec?.queueInputBuffer(
                inputBufferId,
                0,
                byteArray.size,
                extractedPts,
                flag,
            )
        }

        while (true) {
            val outputBufferId: Int = mediaCodec?.dequeueOutputBuffer(bufferInfo, 8) ?: 0
            if (outputBufferId <= 0) break
            mediaCodec?.releaseOutputBuffer(outputBufferId, flag == 0)
        }
    }

    private fun createFormat(
        videoMimeType: String?,
    ): MediaFormat {
        return MediaFormat().apply {
            setString(MediaFormat.KEY_MIME, videoMimeType)
            setInteger(MediaFormat.KEY_BIT_RATE, VIDEO_BITRATE)
            // must be present to configure the decoder, but does not impact the actual frame rate, which is variable
            setInteger(MediaFormat.KEY_FRAME_RATE, 60)
            setInteger(
                MediaFormat.KEY_COLOR_FORMAT,
                MediaCodecInfo.CodecCapabilities.COLOR_FormatSurface
            )
            setInteger(MediaFormat.KEY_COLOR_RANGE, MediaFormat.COLOR_RANGE_LIMITED)
            setInteger(MediaFormat.KEY_PRIORITY, 0)
        }
    }

    fun stop() {
        try {
            mediaCodec?.stop()
        } catch (_: Exception) {
        }

        try {
            mediaCodec?.release()
        } catch (_: Exception) {
        }
        mediaCodec = null

        try {
            videoStreamer.close()
        } catch (_: Exception) {
        }

        try {
            surface?.release()
            surface = null
        } catch (_: Exception) {
        }
    }

    companion object {
        private const val VIDEO_BITRATE = 8000000

        private const val PACKET_FLAG_CONFIG: Long = 1L shl 63
        private const val PACKET_FLAG_KEY_FRAME: Long = 1L shl 62
    }
}