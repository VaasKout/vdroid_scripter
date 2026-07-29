package com.vision.scripter.streaming.impl.blocks.video.ui

import android.view.MotionEvent
import android.view.Surface
import androidx.compose.runtime.Stable
import com.vision.scripter.streaming.impl.screen.main.commandobservers.MenuToVideo
import com.vision.scripter.streaming.impl.screen.main.commandobservers.ScreenToVideo
import com.vision.scripter.streaming.impl.screen.main.commandobservers.VideoToMenu
import com.vision.scripter.streaming.impl.screen.main.commandobservers.VideoToScreen
import com.vision.scripter.ui.CommandFlow
import kotlinx.coroutines.flow.StateFlow

@Stable
interface VideoUiStateHolder {

    val uiStateFlow: StateFlow<VideoUiState>
    val menuCommandsFlow: CommandFlow<VideoToMenu>
    val screenCommandsFlow: CommandFlow<VideoToScreen>

    fun init(serial: String)
    fun onSharedEvent(event: ScreenToVideo)
    fun onSharedEvent(event: MenuToVideo)

    fun onVideoSurfaceCreated(surfaceWidth: Int, surfaceHeight: Int, newSurface: Surface)
    fun onVideoSurfaceDestroyed()
    fun onTouchEvent(viewWidth: Int, viewHeight: Int, event: MotionEvent?)
}
