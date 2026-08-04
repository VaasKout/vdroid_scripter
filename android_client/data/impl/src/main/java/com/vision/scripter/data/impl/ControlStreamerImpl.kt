package com.vision.scripter.data.impl

import android.view.MotionEvent
import com.vision.scripter.data.api.ControlStreamer
import com.vision.scripter.data.api.models.ScreenSizes
import java.io.BufferedOutputStream
import java.io.DataOutputStream
import java.io.IOException
import java.net.InetSocketAddress
import java.net.Socket
import java.nio.ByteBuffer
import javax.inject.Inject
import javax.net.SocketFactory

private const val TYPE_INJECT_TOUCH_EVENT = 2
private const val POINTER_ID_GENERIC_FINGER: Long = -2

class ControlStreamerImpl @Inject constructor() : ControlStreamer {

    companion object {
        private const val SOCKET_TIMEOUT_MS: Int = 3000
    }

    private var output: DataOutputStream? = null
    private var socket: Socket? = null

    override fun initConnection(
        host: String,
        port: Int,
    ): Boolean {
        try {
            socket = SocketFactory.getDefault().createSocket()
            socket?.tcpNoDelay = true
            val address = InetSocketAddress(host, port)
            socket?.connect(address, SOCKET_TIMEOUT_MS)
        } catch (e: IOException) {
            e.printStackTrace()
            return false
        }

        if (socket?.isConnected == false) {
            return false
        }

        try {
            output = DataOutputStream(BufferedOutputStream(socket?.getOutputStream()))
        } catch (e: IOException) {
            e.printStackTrace()
            return false
        }

        return true
    }

    @OptIn(ExperimentalUnsignedTypes::class)
    override fun sendControlData(
        screenSizes: ScreenSizes,
        event: MotionEvent?,
    ): UByteArray? {
        try {
            val buffer = ByteBuffer.allocate(32)
            event ?: return null
            output ?: return null

            val remoteX = event.x.toInt() * screenSizes.remoteWidth / screenSizes.surfaceWidth
            val remoteY = event.y.toInt() * screenSizes.remoteHeight / screenSizes.surfaceHeight

            buffer.put(TYPE_INJECT_TOUCH_EVENT.toByte())
            buffer.put(event.action.toByte())
            buffer.putLong(POINTER_ID_GENERIC_FINGER) // pointerId
            buffer.putInt(remoteX)
            buffer.putInt(remoteY)
            buffer.putShort(screenSizes.remoteWidth.toShort())
            buffer.putShort(screenSizes.remoteHeight.toShort())
            buffer.putShort(event.pressure.toInt().toShort())  // pressure
            buffer.putInt(event.actionButton) // action button
            buffer.putInt(0) // buttons

            output?.write(buffer.array())
            output?.flush()
            return buffer.array().toUByteArray()
        } catch (e: Exception) {
            e.printStackTrace()
        }
        return null
    }

    override fun close() {
        try {
            output?.close()
        } catch (e: Exception) {
            e.printStackTrace()
        }

        try {
            socket?.close()
        } catch (e: Exception) {
            e.printStackTrace()
        }
        output = null
        socket = null
    }
}
