package com.vision.scripter.data.impl

import com.vision.scripter.coroutines.api.DispatchersFactory
import com.vision.scripter.data.api.CvStreamer
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.delay
import kotlinx.coroutines.withContext
import java.io.BufferedOutputStream
import java.io.DataInputStream
import java.io.DataOutputStream
import java.io.IOException
import java.net.InetSocketAddress
import java.net.Socket
import javax.inject.Inject
import javax.net.SocketFactory
import kotlin.time.Duration.Companion.milliseconds

class CvStreamerImpl @Inject constructor(
    private val dispatchersFactory: DispatchersFactory,
) : CvStreamer {

    companion object {
        private const val SOCKET_TIMEOUT_MS: Int = 3000
        private const val MAX_FRAME_SIZE = 1024 * 1024 // 1MB
    }

    @Volatile
    private var socket: Socket? = null

    @Volatile
    private var input: DataInputStream? = null

    @Volatile
    private var output: DataOutputStream? = null

    override suspend fun connect(
        host: String,
        port: Int,
    ): Boolean = withContext(Dispatchers.IO) {
        try {
            socket = SocketFactory.getDefault().createSocket()
            val address = InetSocketAddress(host, port)
            socket?.connect(address, SOCKET_TIMEOUT_MS)
        } catch (e: IOException) {
            e.printStackTrace()
            return@withContext false
        }

        if (socket?.isConnected == false) {
            return@withContext false
        }

        try {
            input = DataInputStream(socket?.getInputStream())
        } catch (e: IOException) {
            e.printStackTrace()
            return@withContext false
        }

        try {
            output = DataOutputStream(BufferedOutputStream(socket?.getOutputStream()))
        } catch (e: IOException) {
            e.printStackTrace()
            return@withContext false
        }

        return@withContext true
    }

    override suspend fun readRectangles(): ByteArray? = withContext(dispatchersFactory.io) {
        try {
            val size = input?.readInt() ?: return@withContext null
            if (size <= 0) return@withContext null
            if (size > MAX_FRAME_SIZE) {
                input?.skipBytes(size)
                return@withContext null
            }
            val buffer = ByteArray(size)
            input?.readFully(buffer)
            buffer
        } catch (e: Exception) {
            e.printStackTrace()
            delay(1000.milliseconds)
            null
        }
    }

    override suspend fun sendCvMode(cvMode: Int): Boolean = withContext(dispatchersFactory.io) {
        try {
            output ?: return@withContext false
            output?.writeInt(cvMode)
            output?.flush()
        } catch (e: Exception) {
            e.printStackTrace()
            return@withContext false
        }
        return@withContext true
    }

    override fun close() {
        try {
            input?.close()
        } catch (e: Exception) {
            e.printStackTrace()
        }

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
        input = null
        output = null
        socket = null
    }
}