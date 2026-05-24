package com.vision.scripter.streaming.impl.state

import android.view.MotionEvent
import android.view.MotionEvent.ACTION_UP
import android.view.Surface
import androidx.core.net.toUri
import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.data.api.ControlStreamer
import com.vision.scripter.data.api.ScripterRepository
import com.vision.scripter.data.api.models.EVENT_ON_TEMPLATE
import com.vision.scripter.data.api.models.EVENT_ON_TEXT
import com.vision.scripter.data.api.models.RectangleWithText
import com.vision.scripter.data.api.models.ScriptStep
import com.vision.scripter.data.api.models.StepEvent
import com.vision.scripter.data.api.models.StreamingData
import com.vision.scripter.data.api.models.TYPE_TEXT
import com.vision.scripter.data.api.models.adjustToClient
import com.vision.scripter.data.api.models.contains
import com.vision.scripter.data.api.models.extractPressEvent
import com.vision.scripter.network.api.ApiResponse
import com.vision.scripter.prefs.api.DataStoreRepository
import com.vision.scripter.streaming.impl.ui.StreamingUiCommand
import com.vision.scripter.streaming.impl.ui.StreamingUiState
import com.vision.scripter.streaming.impl.ui.StreamingUiStateHolder
import com.vision.scripter.streaming.impl.usecases.CvUseCase
import com.vision.scripter.streaming.impl.usecases.VideoUseCase
import com.vision.scripter.ui.CommandFlow
import com.vision.scripter.ui.states.LoadingState
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.launchIn
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.onEach
import kotlinx.coroutines.flow.shareIn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import javax.inject.Inject

@ViewModelScoped
class StreamingInteractor @Inject constructor(
    coroutineScopeFactory: CoroutineScopeFactory,
    private val scripterRepository: ScripterRepository,
    private val dataStoreRepository: DataStoreRepository,
    private val uiStateMapper: StreamingUiStateMapper,
    private val videoUseCase: VideoUseCase,
    private val controlStreamer: ControlStreamer,
    private val cvUseCase: CvUseCase,
) : StreamingUiStateHolder {

    private val coroutineScope: CoroutineScope =
        coroutineScopeFactory.createBackgroundScope("streaming_interactor")

    private val _stateFlow = MutableStateFlow(StreamingState())
    private val stateFlow: SharedFlow<StreamingState> = _stateFlow.asStateFlow()

    private val currentState: StreamingState
        get() = _stateFlow.value

    override val uiStateFlow: SharedFlow<StreamingUiState>
        get() = stateFlow.map(uiStateMapper::map)
            .shareIn(coroutineScope, SharingStarted.WhileSubscribed(), replay = 1)

    override val uiCommandsFlow: CommandFlow<StreamingUiCommand> = CommandFlow(coroutineScope)

    private val mutex = Mutex()

    @Volatile
    private var streamJob: Job? = null

    override fun initArgs(serial: String) {
        _stateFlow.update {
            it.copy(serial = serial)
        }
    }

    init {
        cvUseCase.observe(coroutineScope).onEach { rectangles ->
            _stateFlow.update {
                it.copy(cvRectangles = rectangles)
            }
        }.launchIn(coroutineScope)
    }

    override fun onLoadData(onStart: Boolean) {
        coroutineScope.launch {
            _stateFlow.update {
                it.copy(loadingState = LoadingState.LoadingOnStart)
            }
            val result = getStreamingData()
            if (result is ApiResponse.Error) {
                uiCommandsFlow.tryEmit(StreamingUiCommand.ShowNetworkError)
                _stateFlow.update {
                    it.copy(loadingState = LoadingState.None)
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
            val videoConnected = videoUseCase.initConnection(
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

            val cvConnected = cvUseCase.initConnection(
                host = currentState.streamingHost,
                port = streamingData.cvPort.toInt(),
            )

            val screenSizes = videoUseCase.observeScreenSizes().value
            val connectionEstablished =
                videoConnected && controlConnected && cvConnected && screenSizes != null

            _stateFlow.update {
                it.copy(loadingState = LoadingState.None)
            }

            if (!connectionEstablished) {
                _stateFlow.update {
                    it.copy(streamingData = null)
                }
                return@launch
            }

            launch {
                videoUseCase.decodeFramesInLoop(
                    mimeType = currentState.videoCodec.mimeType,
                )
            }
            cvUseCase.decodeRectanglesInLoop(
                screenSizes = screenSizes,
            )
        }
    }

    override fun onVideoSurfaceDestroyed() {
        closeStreams()
    }

    fun closeStreams() {
        coroutineScope.launch {
            if (streamJob?.isActive == true) {
                streamJob?.cancel()
                streamJob = null
            }
            videoUseCase.stop()
            controlStreamer.close()
            cvUseCase.close()
        }
    }

    override fun onTouchEvent(
        viewWidth: Int,
        viewHeight: Int,
        event: MotionEvent?,
    ) {
        if (event == null) return
        coroutineScope.launch {
            mutex.withLock {
                val menuState = currentState.menuState
                if (menuState is MenuState.SelectingCV) {
                    cvUseCase.selectRectangle(x = event.x.toInt(), y = event.y.toInt())
                    return@launch
                }

                if (
                    menuState is MenuState.TypingText &&
                    menuState.recordingKeyboard &&
                    event.x > 0 &&
                    event.y > 0 &&
                    event.action == ACTION_UP
                ) {
                    val letter = currentState.keyboard.buttons.firstOrNull {
                        it.contains(x = event.x.toInt(), y = event.y.toInt())
                    }?.text ?: return@launch

                    if (letter.isEmpty()) return@launch

                    val updatedTypeText = buildString {
                        append(menuState.typeText)
                        if (letter == SPACE_KEY) append(" ")
                        else append(letter)
                    }

                    _stateFlow.update {
                        it.copy(
                            menuState = menuState.copy(
                                typeText = updatedTypeText
                            )
                        )
                    }
                }

                val screenSizes = videoUseCase.observeScreenSizes().value ?: return@launch
                val bytesArray = controlStreamer.sendControlData(
                    screenSizes = screenSizes,
                    event = event,
                )

                recordBytes(bytesArray)
            }
        }
    }

    var startRecordingTime = 0L
    private fun recordBytes(bytesArray: ByteArray?) {
        val menuState = currentState.menuState
        if (
            bytesArray == null ||
            !(menuState is MenuState.TypingText && menuState.recordingKeyboard ||
                    menuState is MenuState.Recording && menuState.controlRecording)
        ) return

        if (startRecordingTime == 0L) {
            startRecordingTime = System.nanoTime()
        }

        val elapsedMs = (System.nanoTime() - startRecordingTime) / 1_000_000L
        val newStepEvent = StepEvent(
            time = elapsedMs,
            data = bytesArray,
        )

        _stateFlow.update {
            it.copy(
                record = it.record.copy(
                    stepEvents = it.record.stepEvents + newStepEvent
                )
            )
        }
    }

    override fun onScriptModeClicked() {
        coroutineScope.launch {
            _stateFlow.update {
                it.copy(
                    showRecordDialog = true,
                )
            }
        }
    }

    override fun onCvModeClicked() {
        coroutineScope.launch {
            val menuState = currentState.menuState
            if (menuState is MenuState.Recording) {
                _stateFlow.update {
                    val selectMode = it.record.templateSelectMode.incrementOnlyActive()
                    it.copy(
                        menuState = MenuState.SelectingCV(
                            selectMode = selectMode,
                        ),
                    )
                }
                cvUseCase.nextCvMode(newCvMode = CVMode.CV_RECTS)
            }

            if (menuState is MenuState.SelectingCV) {
                val selectMode = menuState.selectMode.increment()
                val cvMode = if (selectMode != CvSelectMode.NONE) {
                    CVMode.CV_RECTS
                } else {
                    CVMode.NO_CV
                }
                _stateFlow.update {
                    it.copy(
                        menuState = menuState.copy(
                            selectMode = selectMode,
                        )
                    )
                }
                cvUseCase.nextCvMode(newCvMode = cvMode)
                if (selectMode == CvSelectMode.NONE) cvUseCase.disableSelection()
            }

            if (menuState is MenuState.Usual) {
                val newCVMode = menuState.cvMode.increment()
                _stateFlow.update {
                    it.copy(
                        menuState = menuState.copy(
                            cvMode = newCVMode,
                        )
                    )
                }
                cvUseCase.nextCvMode(newCvMode = newCVMode)
            }
        }
    }

    override fun onRecordingClicked() {
        coroutineScope.launch {
            val mode = currentState.menuState
            when {
                mode is MenuState.Recording -> {
                    _stateFlow.update {
                        it.copy(
                            menuState = mode.copy(
                                controlRecording = !mode.controlRecording,
                            )
                        )
                    }
                }

                mode is MenuState.TypingText -> {
                    _stateFlow.update {
                        it.copy(
                            menuState = mode.copy(
                                recordingKeyboard = !mode.recordingKeyboard,
                                typeText = "",
                            )
                        )
                    }
                }
            }
        }
    }

    override fun onExpandClicked() {
        val menuState = currentState.menuState
        if (menuState is MenuState.Usual) {
            _stateFlow.update {
                it.copy(
                    menuState = menuState.copy(expanded = !menuState.expanded)
                )
            }
        }
    }

    override fun onSaveClicked() {
        coroutineScope.launch {
            val menuState = currentState.menuState
            if (menuState is MenuState.SelectingCV) {
                _stateFlow.update {
                    it.copy(
                        menuState = MenuState.Recording(
                            templateSelectMode = menuState.selectMode,
                        ),
                        record = it.record.copy(
                            templateSelectMode = menuState.selectMode,
                        )
                    )
                }
                val screenSizes = videoUseCase.observeScreenSizes().value ?: return@launch
                cvUseCase.saveSelectedRectangle(
                    serial = currentState.serial,
                    screenSizes = screenSizes,
                )
                cvUseCase.nextCvMode(CVMode.NO_CV)
                return@launch
            }

            if (menuState is MenuState.SelectingText) {
                _stateFlow.update {
                    it.copy(
                        menuState = MenuState.Recording(
                            textSelectMode = menuState.selectMode,
                        ),
                        record = it.record.copy(
                            text = menuState.text,
                            textSelectMode = menuState.selectMode,
                        )
                    )
                }
                cvUseCase.nextCvMode(CVMode.NO_CV)
                return@launch
            }

            if (menuState is MenuState.TypingText) {
                _stateFlow.update {
                    val pressEvent = it.record.stepEvents.extractPressEvent()

                    it.copy(
                        menuState = MenuState.Recording(
                            typeText = menuState.typeText.isNotEmpty(),
                        ),
                        record = it.record.copy(
                            text = menuState.typeText,
                            typeText = menuState.typeText.isNotEmpty(),
                            stepEvents = pressEvent,
                        )
                    )
                }
                cvUseCase.nextCvMode(CVMode.NO_CV)
                return@launch
            }

            val record = currentState.record
            val success = scripterRepository.saveScriptStep(
                serial = currentState.serial,
                name = record.recordName,
                step = ScriptStep(
                    events = record.stepEvents,
                    flags = getFlags(),
                    text = record.text,
                    locale = record.locale,
                ),
            )

            if (!success) return@launch
            uiCommandsFlow.tryEmit(StreamingUiCommand.ShowStepSavedSnackbar)
            startRecordingTime = 0L
            cvUseCase.nextCvMode(CVMode.NO_CV)
            cvUseCase.clearAllRectangles()
            _stateFlow.update {
                it.copy(
                    menuState = MenuState.Recording(),
                    record = it.record.clearStep(),
                )
            }
        }
    }

    private fun getFlags(): Int {
        val record = currentState.record
        return when {
            record.typeText -> TYPE_TEXT
            record.templateSelectMode == CvSelectMode.APPLY_EVENT -> EVENT_ON_TEMPLATE
            record.textSelectMode == CvSelectMode.APPLY_EVENT -> EVENT_ON_TEXT
            else -> 0
        }
    }

    override fun onCancelClicked() {
        coroutineScope.launch {
            val menuState = currentState.menuState
            if (menuState is MenuState.Recording) {
                cvUseCase.clearAllRectangles()
                _stateFlow.update {
                    it.copy(
                        menuState = MenuState.Usual(
                            expanded = true,
                        )
                    )
                }
            } else {
                _stateFlow.update {
                    it.copy(
                        menuState = MenuState.Recording(
                            templateSelectMode = it.record.templateSelectMode,
                            textSelectMode = it.record.textSelectMode,
                            typeText = it.record.typeText,
                        )
                    )
                }
            }

            cvUseCase.nextCvMode(CVMode.NO_CV)
        }
    }

    override fun onSavedRecordName(name: String) {
        coroutineScope.launch {
            _stateFlow.update {
                it.copy(
                    record = it.record.copy(recordName = name),
                    showRecordDialog = false,
                    menuState = MenuState.Recording(),
                )
            }
            cvUseCase.nextCvMode(CVMode.NO_CV)
        }
    }

    override fun onSaveLocale(locale: String) {
        coroutineScope.launch {
            _stateFlow.update {
                val updatedMenuState = when (it.menuState) {
                    is MenuState.Recording -> MenuState.TypingText()
                    is MenuState.Usual -> it.menuState.copy(keyboardLoading = true)
                    else -> it.menuState
                }
                it.copy(
                    record = it.record.copy(locale = locale),
                    showKeyboardDialog = false,
                    menuState = updatedMenuState,
                )
            }

            getOrResetKeyboard()
        }
    }

    override fun onDialogDismissed() {
        val menuState = currentState.menuState
        _stateFlow.update {
            it.copy(
                showRecordDialog = false,
                showTextDialog = false,
                showKeyboardDialog = false,
            )
        }

        if (menuState is MenuState.Recording) {
            _stateFlow.update {
                it.copy(
                    menuState = MenuState.Recording(
                        templateSelectMode = it.record.templateSelectMode,
                        textSelectMode = it.record.textSelectMode,
                        typeText = it.record.typeText,
                    ),
                )
            }
        }
    }

    override fun onTextModeClicked() {
        coroutineScope.launch {
            val menuState = currentState.menuState
            if (menuState is MenuState.Recording) {
                _stateFlow.update {
                    it.copy(showTextDialog = true)
                }
                return@launch
            }

            if (menuState is MenuState.SelectingText) {
                val newTextMode = menuState.selectMode.increment()
                _stateFlow.update {
                    it.copy(
                        menuState = menuState.copy(
                            selectMode = newTextMode,
                        )
                    )
                }
                if (newTextMode != CvSelectMode.NONE) {
                    cvUseCase.selectAll()
                    return@launch
                }
                cvUseCase.disableSelection()
            }
        }
    }

    override fun onTryToFindText(text: String, locale: String) {
        coroutineScope.launch {
            _stateFlow.update {
                it.copy(
                    showTextDialog = false,
                )
            }

            val screenSizes = videoUseCase.observeScreenSizes().value ?: return@launch
            val found = cvUseCase.findTextRectangles(
                serial = currentState.serial,
                text = text.trim(),
                locale = locale,
                screenSizes = screenSizes,
            )

            if (!found) {
                uiCommandsFlow.tryEmit(StreamingUiCommand.ShowNetworkError)
                return@launch
            }

            val menuState = currentState.menuState
            if (menuState is MenuState.Recording) {
                _stateFlow.update {
                    it.copy(
                        menuState = MenuState.SelectingText(
                            selectMode = menuState.textSelectMode.increment(),
                            text = text.trim(),
                        )
                    )
                }
            }
        }
    }

    override fun onKeyboardClicked() {
        coroutineScope.launch {
            val menuState = currentState.menuState
            if (menuState is MenuState.Usual && menuState.showKeyboardButtons) {
                _stateFlow.update {
                    it.copy(
                        menuState = menuState.copy(
                            showKeyboardButtons = false,
                            keyboardLoading = false,
                        )
                    )
                }
                return@launch
            }
            _stateFlow.update {
                it.copy(
                    showKeyboardDialog = true,
                )
            }
        }
    }

    override fun onKeyboardInitClicked() {
        coroutineScope.launch {
            val menuState = currentState.menuState
            val updatedMenuState = when (menuState) {
                is MenuState.Usual -> {
                    menuState.copy(keyboardLoading = true)
                }

                is MenuState.TypingText -> {
                    menuState.copy(isLoadingKeyboard = true)
                }

                else -> menuState
            }
            _stateFlow.update {
                it.copy(
                    menuState = updatedMenuState,
                )
            }

            val result = scripterRepository.resetKeyboard(
                serial = currentState.serial,
                locale = currentState.record.locale,
            )

            if (result is ApiResponse.Success) {
                setupKeyboardRects(result.data)
            }
        }
    }

    private suspend fun getOrResetKeyboard() {
        val keyboardResult = scripterRepository.getKeyboard(
            serial = currentState.serial,
            locale = currentState.record.locale,
        )
        if (keyboardResult is ApiResponse.Success && keyboardResult.data.isNotEmpty()) {
            setupKeyboardRects(keyboardResult.data)
            return
        }

        val resetResult = scripterRepository.resetKeyboard(
            serial = currentState.serial,
            locale = currentState.record.locale,
        )
        if (resetResult is ApiResponse.Success) {
            setupKeyboardRects(resetResult.data)
            return
        }
        showKeyboardError()
    }

    private fun setupKeyboardRects(data: List<RectangleWithText>) {
        val screenSizes = videoUseCase.observeScreenSizes().value ?: return
        val buttons = data.mapNotNull {
            val rectangle = it.rectangle ?: return@mapNotNull null
            it.copy(rectangle = rectangle.adjustToClient(screenSizes))
        }

        val menuState = currentState.menuState
        val updatedMenuState = when (menuState) {
            is MenuState.Usual -> {
                menuState.copy(
                    keyboardLoading = false,
                    showKeyboardButtons = buttons.isNotEmpty()
                )
            }

            is MenuState.TypingText -> {
                menuState.copy(isLoadingKeyboard = false)
            }

            else -> menuState
        }

        _stateFlow.update {
            it.copy(
                keyboard = it.keyboard.copy(buttons = buttons),
                menuState = updatedMenuState,
            )
        }
    }

    private fun showKeyboardError() {
        val menuState = currentState.menuState
        val updatedMenuState = when (menuState) {
            is MenuState.Usual -> {
                menuState.copy(keyboardLoading = false)
            }

            is MenuState.TypingText -> {
                menuState.copy(isLoadingKeyboard = false)
            }

            else -> menuState
        }

        _stateFlow.update {
            it.copy(
                menuState = updatedMenuState,
            )
        }

        uiCommandsFlow.tryEmit(StreamingUiCommand.ShowNetworkError)
    }

    suspend fun getStreamingData(): ApiResponse<StreamingData> {
        val result = scripterRepository.startSockets(currentState.serial)
        when (result) {
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
                uiCommandsFlow.tryEmit(StreamingUiCommand.ShowNetworkError)
            }
        }
        return result
    }
}