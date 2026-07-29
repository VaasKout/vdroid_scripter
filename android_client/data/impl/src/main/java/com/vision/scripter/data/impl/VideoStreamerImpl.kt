package com.vision.scripter.data.impl

import com.vision.scripter.data.api.VideoStreamer
import java.io.BufferedInputStream
import java.io.DataInputStream
import java.io.IOException
import java.net.InetSocketAddress
import java.net.Socket
import javax.inject.Inject
import javax.net.SocketFactory

class VideoStreamerImpl @Inject constructor() : VideoStreamer {

    companion object {
        private const val SOCKET_TIMEOUT_MS: Int = 3000
    }

    private var socket: Socket? = null
    private var input: DataInputStream? = null

    override fun connect(
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
            input = DataInputStream(BufferedInputStream(socket?.getInputStream()))
        } catch (e: IOException) {
            e.printStackTrace()
            return false
        }

        return true
    }

    override fun readVideoHeader(): Pair<Int, Int> {
        val width = input?.readInt() ?: 0
        val height = input?.readInt() ?: 0
        return Pair(width, height)
    }

    override fun readFrameMeta(): Pair<Long, Int> {
        val ptsAndFlags = input?.readLong() ?: 0L
        val packetSize = input?.readInt() ?: 0
        return Pair(ptsAndFlags, packetSize)
    }

    override fun readFrame(size: Int): ByteArray {
        val buffer = ByteArray(size)
        input?.readFully(buffer)
        return buffer
    }


    override fun close() {
        try {
            input?.close()
        } catch (e: Exception) {
            e.printStackTrace()
        }

        try {
            socket?.close()
        } catch (e: Exception) {
            e.printStackTrace()
        }
        input = null
        socket = null
    }
}