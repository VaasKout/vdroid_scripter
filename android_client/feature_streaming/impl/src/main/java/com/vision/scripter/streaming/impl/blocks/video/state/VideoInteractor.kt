package com.vision.scripter.streaming.impl.blocks.video.state

import android.view.MotionEvent
import android.view.MotionEvent.ACTION_DOWN
import android.view.Surface
import androidx.core.net.toUri
import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.data.api.ControlStreamer
import com.vision.scripter.data.api.ScripterDataSource
import com.vision.scripter.data.api.models.Parameter
import com.vision.scripter.data.api.models.adjustToServer
import com.vision.scripter.network.api.ApiResponse
import com.vision.scripter.prefs.api.DataStoreRepository
import com.vision.scripter.streaming.impl.blocks.video.ui.VideoUiState
import com.vision.scripter.streaming.impl.blocks.video.ui.VideoUiStateHolder
import com.vision.scripter.streaming.impl.data.CvStreamerRepository
import com.vision.scripter.streaming.impl.data.KeyboardRepository
import com.vision.scripter.streaming.impl.data.RecordRepository
import com.vision.scripter.streaming.impl.data.VideoStreamerRepository
import com.vision.scripter.streaming.impl.screen.commandobservers.MenuToVideo
import com.vision.scripter.streaming.impl.screen.commandobservers.ScreenToVideo
import com.vision.scripter.streaming.impl.screen.commandobservers.VideoToMenu
import com.vision.scripter.streaming.impl.screen.commandobservers.VideoToScreen
import com.vision.scripter.streaming.impl.screen.state.CVMode
import com.vision.scripter.streaming.impl.screen.state.KeyboardMode
import com.vision.scripter.streaming.impl.screen.state.TEMPLATE
import com.vision.scripter.streaming.impl.screen.state.TEXT
import com.vision.scripter.streaming.impl.screen.state.YOLO_CLASS
import com.vision.scripter.streaming.impl.screen.state.toType
import com.vision.scripter.ui.CommandFlow
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

    override val menuCommandsFlow: CommandFlow<VideoToMenu> = CommandFlow(coroutineScope)
    override val screenCommandsFlow: CommandFlow<VideoToScreen> = CommandFlow(coroutineScope)

    private val mutex = Mutex()
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
        onLoadData()
    }

    override fun onSharedEvent(event: ScreenToVideo) {
        when (event) {
            is ScreenToVideo.Refresh -> onLoadData()
        }
    }

    override fun onSharedEvent(event: MenuToVideo) {
        when (event) {
            is MenuToVideo.SaveClicked -> onSaveClicked()
            is MenuToVideo.OnCvModeClicked -> nextCvMode(event.nextMode, event.recording)
            is MenuToVideo.OnTextModeClicked -> nextCvMode(CVMode.NO_CV, false)
            is MenuToVideo.CancelClicked -> onCancelClicked()
            is MenuToVideo.FindText -> findTextClicked(event.text, event.locale)
            is MenuToVideo.KeyboardStateChanged -> changeKeyboardState(event.keyboardMode)
            is MenuToVideo.KeyboardButtonEdited -> editKeyboardKey(
                oldKey = event.oldKey,
                newKey = event.newKey,
            )

            is MenuToVideo.KeyboardLocaleSaved -> openKeyboard(locale = event.locale)
            is MenuToVideo.RecordNameSaved -> saveRecordNameClicked(event.name)
            is MenuToVideo.TimeoutSaved -> saveTimeoutClicked(event.timeout)
            is MenuToVideo.OnRecordingClicked -> onRecordingClicked(event.isControlRecording)
        }
    }

    private fun onLoadData() {
        coroutineScope.launch {
            when (val result = scripterDataSource.startSockets(currentState.serial)) {
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
                    screenCommandsFlow.tryEmit(VideoToScreen.ShowNetworkError)
                }
            }
        }
    }

    private fun nextCvMode(cvMode: CVMode, recording: Boolean) {
        coroutineScope.launch {
            if (recording) {
                val type = cvMode.toType()
                if (type.isNotEmpty()) {
                    _stateFlow.update {
                        it.copy(
                            tmpParam = Parameter(type = type),
                        )
                    }
                }
            }
            if (cvMode == CVMode.NO_CV) {
                cvRepository.restoreSelectedRectangles()
            } else {
                cvRepository.snapshotSelectedRectangles()
            }
            cvRepository.nextCvMode(cvMode)
        }
    }

    private fun onSaveClicked() {
        coroutineScope.launch {
            mutex.withLock {
                if (currentState.keyboardButtons.isNotEmpty()) {
                    saveTypedText()
                    return@launch
                }

                val parameter = currentState.tmpParam
                if (parameter == null) {
                    saveScript()
                    return@launch
                }

                if (parameter.type == TEMPLATE) {
                    val screenSizes = currentState.screenSizes ?: return@launch
                    cvRepository.nextCvMode(CVMode.NO_CV)
                    val templateName = "${currentState.record.params.size + 1}"
                    cvRepository.saveSelectedRectangle(
                        serial = currentState.serial,
                        node = currentState.record.node,
                        name = currentState.record.name,
                        value = templateName,
                        screenSizes = screenSizes,
                    )
                    addParam(parameter.copy(value = templateName))
                    return@launch
                }

                if (parameter.type == YOLO_CLASS) {
                    val label =
                        cvRepository.observeSelectedRectangles().value.firstOrNull()?.label.orEmpty()
                    cvRepository.nextCvMode(CVMode.NO_CV)
                    addParam(parameter.copy(value = label))
                    return@launch
                }

                if (parameter.value.isNotEmpty()) {
                    addParam(parameter)
                }
            }
        }
    }

    private fun onRecordingClicked(isControlRecording: Boolean) {
        recordRepository.setControlRecording(isControlRecording)
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
                screenCommandsFlow.tryEmit(VideoToScreen.ShowNetworkError)
                return@launch
            }

            screenCommandsFlow.tryEmit(VideoToScreen.SuccessLoading)

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

    override fun onTouchEvent(
        viewWidth: Int,
        viewHeight: Int,
        event: MotionEvent?,
    ) {
        if (event == null) return
        val screenSizes = currentState.screenSizes ?: return
        val record = currentState.record
        val tmpParam = currentState.tmpParam

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

                    if (newButton != null) {
                        menuCommandsFlow.tryEmit(
                            VideoToMenu.SelectKeyboardKey(newButton),
                        )
                        return@launch
                    }

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

    private fun onCancelClicked() {
        coroutineScope.launch {
            mutex.withLock {
                val parameter = currentState.tmpParam
                recordRepository.clear(parameter != null)
                dropState()
            }
        }
    }

    private fun findTextClicked(text: String, locale: String) {
        coroutineScope.launch {
            mutex.withLock {
                val screenSizes = currentState.screenSizes ?: return@launch
                val found = cvRepository.findTextRectangles(
                    serial = currentState.serial,
                    text = text,
                    locale = locale,
                    screenSizes = screenSizes,
                )
                if (!found) {
                    screenCommandsFlow.tryEmit(VideoToScreen.ShowNetworkError)
                    return@launch
                }
                _stateFlow.update {
                    it.copy(tmpParam = Parameter(type = TEXT, value = text, locale = locale))
                }
            }
        }
    }

    private fun changeKeyboardState(keyboardMode: KeyboardMode) {
        coroutineScope.launch {
            mutex.withLock {
                keyboardRepository.updateKeyboardState(keyboardMode)
                cvRepository.clearSelectedRectangles()
                if (keyboardMode == KeyboardMode.ADD_NEW) {
                    cvRepository.nextCvMode(CVMode.CV_RECTS)
                    return@withLock
                }
                cvRepository.nextCvMode(CVMode.NO_CV)
            }
        }
    }

    private fun editKeyboardKey(oldKey: String, newKey: String) {
        coroutineScope.launch {
            mutex.withLock {
                val screenSizes = currentState.screenSizes ?: return@launch
                val tmpZone =
                    cvRepository.observeSelectedRectangles().value.firstOrNull() ?: return@launch
                val success = keyboardRepository.editKeyboardKey(
                    serial = currentState.serial,
                    oldName = oldKey,
                    newName = newKey,
                    rectangle = tmpZone.adjustToServer(screenSizes),
                )
                cvRepository.clearSelectedRectangles()
                if (success) {
                    getOrResetKeyboard()
                    return@launch
                }
                screenCommandsFlow.tryEmit(VideoToScreen.ShowNetworkError)
            }
        }
    }

    private fun openKeyboard(locale: String) {
        coroutineScope.launch {
            mutex.withLock {
                keyboardRepository.updateKeyboardLocale(locale)
                getOrResetKeyboard()
            }
        }
    }

    private fun saveTimeoutClicked(timeout: Int) {
        recordRepository.updateTimeout(timeout)
    }

    private fun saveRecordNameClicked(name: String) {
        recordRepository.updateName(name)
    }

    private suspend fun saveTypedText() {
        val param = keyboardRepository.extractParameter()
        if (param == null) {
            recordRepository.clear()
            dropState()
            return
        }
        addParam(param)
    }

    private suspend fun dropState() {
        _stateFlow.update {
            it.copy(tmpParam = null)
        }
        cvRepository.clearSelectedRectangles()
        cvRepository.nextCvMode(CVMode.NO_CV)
        keyboardRepository.clear()
    }

    private fun addParam(param: Parameter) {
        coroutineScope.launch {
            recordRepository.addParam(param)
            dropState()
        }
    }

    private suspend fun saveScript() {
        val success = recordRepository.saveScript()
        if (!success) return

        screenCommandsFlow.tryEmit(VideoToScreen.ShowScriptSavedSnackbar)
        recordRepository.clear()
        dropState()
    }

    private suspend fun getOrResetKeyboard() {
        val screenSizes = currentState.screenSizes ?: return
        menuCommandsFlow.tryEmit(VideoToMenu.SetKeyboardLoading(true))
        val loaded = keyboardRepository.getOrResetKeyboard(
            serial = currentState.serial,
            screenSizes = screenSizes,
        )
        menuCommandsFlow.tryEmit(VideoToMenu.SetKeyboardLoading(false))
        if (!loaded) screenCommandsFlow.tryEmit(VideoToScreen.ShowNetworkError)
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
