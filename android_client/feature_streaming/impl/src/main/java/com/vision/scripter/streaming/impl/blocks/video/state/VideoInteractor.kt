package com.vision.scripter.streaming.impl.blocks.video.state

import android.view.MotionEvent
import android.view.MotionEvent.ACTION_DOWN
import android.view.Surface
import androidx.core.net.toUri
import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.data.api.ControlStreamer
import com.vision.scripter.data.api.ScripterDataSource
import com.vision.scripter.network.api.ApiResponse
import com.vision.scripter.prefs.api.DataStoreRepository
import com.vision.scripter.streaming.impl.blocks.video.ui.VideoUiState
import com.vision.scripter.streaming.impl.blocks.video.ui.VideoUiStateHolder
import com.vision.scripter.streaming.impl.data.CvStreamerRepository
import com.vision.scripter.streaming.impl.data.KeyboardRepository
import com.vision.scripter.streaming.impl.data.RecordRepository
import com.vision.scripter.streaming.impl.screen.StreamingEvent
import com.vision.scripter.streaming.impl.screen.StreamingEventsHolder
import com.vision.scripter.streaming.impl.data.VideoStreamerRepository
import com.vision.scripter.streaming.impl.screen.state.TEMPLATE
import com.vision.scripter.streaming.impl.screen.state.YOLO_CLASS
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.cancel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.launchIn
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import javax.inject.Inject

@ViewModelScoped
class VideoInteractor @Inject constructor(
    coroutineScopeFactory: CoroutineScopeFactory,
    uiStateMapper: VideoUiStateMapper,
    private val scripterDataSource: ScripterDataSource,
    private val dataStoreRepository: DataStoreRepository,
    private val videoRepository: VideoStreamerRepository,
    private val controlStreamer: ControlStreamer,
    private val cvRepository: CvStreamerRepository,
    private val keyboardRepository: KeyboardRepository,
    private val recordRepository: RecordRepository,
    private val eventsHolder: StreamingEventsHolder,
) : VideoUiStateHolder {

    private val coroutineScope: CoroutineScope =
        coroutineScopeFactory.createBackgroundScope("video_interactor")

    private val _stateFlow = MutableStateFlow(VideoState())
    private val stateFlow = _stateFlow.asStateFlow()

    private val currentState: VideoState
        get() = _stateFlow.value

    override val uiStateFlow: StateFlow<VideoUiState> = stateFlow
        .map(uiStateMapper::map)
        .stateIn(coroutineScope, SharingStarted.Eagerly, VideoUiState())

    private val touchMutex = Mutex()

    @Volatile
    private var streamJob: Job? = null

    init {
        startReactiveStreams()
    }

    private fun startReactiveStreams() {
        combine(
            cvRepository.observeRectangles(),
            cvRepository.observeSelectedRectangles(),
            videoRepository.observeScreenSizes(),
            keyboardRepository.observeKeyboardButtons(),
            recordRepository.observeRecord(),
        ) { rectangles, selectedRects, screenSizes, keyboardButtons, record ->
            _stateFlow.update {
                it.copy(
                    cvRectangles = rectangles,
                    selectedRectangles = selectedRects,
                    screenSizes = screenSizes,
                    keyboardButtons = keyboardButtons,
                    record = record,
                )
            }
        }.launchIn(coroutineScope)
    }

    override fun init(serial: String) {
        _stateFlow.update {
            it.copy(serial = serial)
        }
        onRefresh()
    }

    override fun onRefresh() {
        coroutineScope.launch {
            when (val result = scripterDataSource.startSession(currentState.serial)) {
                is ApiResponse.Success -> {
                    val fullServerUri = dataStoreRepository.getServerUrl().toUri()
                    _stateFlow.update {
                        it.copy(
                            streamingHost = fullServerUri.host.orEmpty(),
                            streamingData = result.data
                        )
                    }
                }

                is ApiResponse.Error -> {
                    eventsHolder.sendEvent(StreamingEvent.ShowNetworkError)
                }
            }
        }
    }

    override fun onVideoSurfaceCreated(
        surfaceWidth: Int,
        surfaceHeight: Int,
        newSurface: Surface,
    ) {
        streamJob = coroutineScope.launch {
            val streamingData = currentState.streamingData ?: return@launch
            val videoConnected = videoRepository.initConnection(
                host = currentState.streamingHost,
                port = streamingData.videoPort.toInt(),
                newSurface = newSurface,
                surfaceWidth = surfaceWidth,
                surfaceHeight = surfaceHeight,
            )

            val controlConnected = controlStreamer.initConnection(
                host = currentState.streamingHost,
                port = streamingData.controlPort.toInt(),
            )

            val cvConnected = cvRepository.initConnection(
                host = currentState.streamingHost,
                port = streamingData.cvPort.toInt(),
            )

            val screenSizes = currentState.screenSizes
            val connectionEstablished =
                videoConnected && controlConnected && cvConnected && screenSizes != null

            if (!connectionEstablished) {
                _stateFlow.update {
                    it.copy(streamingData = null)
                }
                eventsHolder.sendEvent(StreamingEvent.ShowNetworkError)
                return@launch
            }

            eventsHolder.sendEvent(StreamingEvent.SuccessLoading)

            launch {
                videoRepository.decodeFramesInLoop(
                    mimeType = currentState.videoCodec.mimeType,
                )
            }
            cvRepository.decodeRectanglesInLoop(
                screenSizes = screenSizes,
            )
        }
    }

    override fun onVideoSurfaceDestroyed() {
        closeStreams()
    }

    @OptIn(ExperimentalUnsignedTypes::class)
    override fun onTouchEvent(
        viewWidth: Int,
        viewHeight: Int,
        event: MotionEvent?,
    ) {
        if (event == null) return
        val screenSizes = currentState.screenSizes ?: return
        val record = currentState.record
        val tmpParam = record.tmpParam

        coroutineScope.launch {
            touchMutex.withLock {
                try {
                    if (tmpParam?.type == TEMPLATE || tmpParam?.type == YOLO_CLASS) {
                        if (event.action == ACTION_DOWN) {
                            cvRepository.selectRectangle(x = event.x.toInt(), y = event.y.toInt())
                        }
                        return@launch
                    }

                    val newButton = keyboardRepository.handleTouchEvent(
                        event,
                        record.name,
                    )

                    if (newButton != null) return@launch

                    val bytesArray = controlStreamer.sendControlData(
                        screenSizes = screenSizes,
                        event = event,
                    )

                    recordRepository.recordBytes(bytesArray)
                } finally {
                    event.recycle()
                }
            }
        }
    }

    fun closeStreams() {
        if (streamJob?.isActive == true) {
            streamJob?.cancel()
            streamJob = null
        }
        videoRepository.stop()
        controlStreamer.close()
        cvRepository.close()
    }

    fun clear() {
        coroutineScope.cancel()
    }
}
