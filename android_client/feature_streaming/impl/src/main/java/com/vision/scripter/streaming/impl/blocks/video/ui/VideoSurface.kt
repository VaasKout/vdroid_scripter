package com.vision.scripter.streaming.impl.blocks.video.ui

import android.annotation.SuppressLint
import android.graphics.SurfaceTexture
import android.view.MotionEvent
import android.view.Surface
import android.view.TextureView
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.viewinterop.AndroidView

@SuppressLint("ClickableViewAccessibility")
@Composable
fun VideoSurface(
    modifier: Modifier = Modifier,
    onSurfaceCreated: (Int, Int, Surface) -> Unit,
    onSurfaceDestroyed: () -> Unit,
    onTouch: (viewWidth: Int, viewHeight: Int, event: MotionEvent?) -> Unit,
) {
    AndroidView(
        modifier = modifier,
        factory = { context ->
            TextureView(context).apply {
                surfaceTextureListener = object : TextureView.SurfaceTextureListener {
                    override fun onSurfaceTextureAvailable(
                        surfaceTexture: SurfaceTexture,
                        width: Int,
                        height: Int
                    ) {
                        surfaceTexture.setDefaultBufferSize(width, height)
                        val surface = Surface(surfaceTexture)
                        onSurfaceCreated(width, height, surface)
                    }

                    override fun onSurfaceTextureDestroyed(surfaceTexture: SurfaceTexture): Boolean {
                        onSurfaceDestroyed()
                        return true
                    }

                    override fun onSurfaceTextureSizeChanged(
                        surface: SurfaceTexture,
                        width: Int,
                        height: Int
                    ) {
                    }

                    override fun onSurfaceTextureUpdated(surface: SurfaceTexture) {}

                }
                setOnTouchListener { view, event ->
                    onTouch(
                        view.width,
                        view.height,
                        MotionEvent.obtain(event),
                    )
                    true
                }
            }
        }
    )
}
