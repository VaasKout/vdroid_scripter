package com.vision.scripter.streaming.impl.state

import android.view.MotionEvent
import android.view.MotionEvent.ACTION_DOWN
import android.view.MotionEvent.ACTION_UP
import android.view.Surface
import androidx.core.net.toUri
import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.data.api.ControlStreamer
import com.vision.scripter.data.api.ScripterRepository
import com.vision.scripter.data.api.models.RectangleWithText
import com.vision.scripter.data.api.models.ScriptStep
import com.vision.scripter.data.api.models.StepEvent
import com.vision.scripter.data.api.models.StreamingData
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
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.launchIn
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
    private val menuInteractor: MenuInteractor,
) : StreamingUiStateHolder {

    private val coroutineScope: CoroutineScope =
        coroutineScopeFactory.createBackgroundScope("streaming_interactor")

    private val _stateFlow = MutableStateFlow(StreamingState())
    private val stateFlow: SharedFlow<StreamingState> = _stateFlow.asStateFlow()

    private val currentState: StreamingState
        get() = _stateFlow.value

    override val uiStateFlow: SharedFlow<StreamingUiState>
        get() = combine(
            stateFlow,
            menuInteractor.observeMenuState(),
            menuInteractor.observeDialogState(),
        ) { streamingState, menuState, dialogState ->
            uiStateMapper.map(
                state = streamingState,
                menuState = menuState,
                dialogState = dialogState,
            )
        }.shareIn(coroutineScope, SharingStarted.WhileSubscribed(), replay = 1)

    override val uiCommandsFlow: CommandFlow<StreamingUiCommand> = CommandFlow(coroutineScope)

    private val mutex = Mutex()

    @Volatile
    private var streamJob: Job? = null

    private var startRecordingTime = 0L

    override fun initArgs(serial: String) {
        _stateFlow.update {
            it.copy(serial = serial)
        }
    }

    init {
        cvUseCase.observeRectangles(coroutineScope).onEach { rectangles ->
            _stateFlow.update {
                it.copy(cvRectangles = rectangles)
            }
        }.launchIn(coroutineScope)

        cvUseCase.observeSelectedRectangles().onEach { selected ->
            _stateFlow.update {
                it.copy(selectedRectangles = selected)
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
                val menuState = menuInteractor.observeMenuState().value
                if (menuState is MenuState.SelectingCV) {
                    if (event.action == ACTION_DOWN) {
                        cvUseCase.selectRectangle(x = event.x.toInt(), y = event.y.toInt())
                    }
                    return@launch
                }

                selectExistingKeyboardButton(event)
                selectNewKeyboardButton(event)
                if (openEditKeyboardDialog()) return@launch

                if (menuState is MenuState.Keyboard && menuState.recordingKeyboard) {
                    val letter = currentState.keyboard.buttons.firstOrNull {
                        it.contains(x = event.x.toInt(), y = event.y.toInt())
                    }?.text ?: return@launch

                    if (letter.isEmpty()) return@launch
                    if (event.action == ACTION_UP) menuInteractor.appendTypedLetter(letter)
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

    private fun openEditKeyboardDialog(): Boolean {
        val menuState = menuInteractor.observeMenuState().value
        val selectedRects = cvUseCase.observeSelectedRectangles().value
        if (
            menuState is MenuState.Keyboard &&
            (menuState.editing || menuState.showCvRectangles) &&
            selectedRects.isNotEmpty()
        ) {
            menuInteractor.onEditKeyboardRectangleSelected()
            return true
        }
        return false
    }

    private fun selectExistingKeyboardButton(event: MotionEvent) {
        val menuState = menuInteractor.observeMenuState().value
        if (
            menuState is MenuState.Keyboard &&
            menuState.editing
        ) {
            val button = currentState.keyboard.buttons.firstOrNull {
                it.contains(x = event.x.toInt(), y = event.y.toInt())
            } ?: return
            menuInteractor.setKeyboardOldKey(button.text)
            cvUseCase.setSelectedRectangle(button.rectangle)
        }
    }

    private fun selectNewKeyboardButton(event: MotionEvent) {
        val menuState = menuInteractor.observeMenuState().value
        if (
            menuState is MenuState.Keyboard &&
            menuState.showCvRectangles
        ) {
            menuInteractor.setKeyboardOldKey("")
            cvUseCase.selectRectangle(x = event.x.toInt(), y = event.y.toInt())
        }
    }

    private fun recordBytes(bytesArray: ByteArray?) {
        val menuState = menuInteractor.observeMenuState().value
        if (
            bytesArray == null ||
            !(menuState is MenuState.Keyboard && menuState.recordingKeyboard ||
                    menuState is MenuState.Recording && menuState.controlRecording)
        ) return

        val elapsedMs = if (startRecordingTime == 0L) {
            startRecordingTime = System.nanoTime()
            0L
        } else {
            (System.nanoTime() - startRecordingTime) / 1_000_000L
        }

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
        menuInteractor.onScriptModeClicked()
    }

    override fun onExpandClicked() {
        menuInteractor.onExpandClicked()
    }

    override fun onRecordingClicked() {
        menuInteractor.onRecordingClicked()
    }

    override fun onKeyboardClicked() {
        menuInteractor.onKeyboardClicked()
    }

    override fun onDialogDismissed() {
        if (menuInteractor.observeMenuState().value is MenuState.Keyboard) {
            cvUseCase.clearSelectedRectangles()
        }
        menuInteractor.onDialogDismissed(currentState.record)
    }

    override fun onCvModeClicked() {
        coroutineScope.launch {
            val enteringSelection = menuInteractor.observeMenuState().value is MenuState.Recording
            val action = menuInteractor.onCvModeClicked() ?: return@launch
            if (enteringSelection) {
                cvUseCase.snapshotSelectedRectangles()
            }
            cvUseCase.nextCvMode(action.newCvMode)
            if (action.disableSelection) {
                cvUseCase.clearSelectedRectangles()
            }
        }
    }

    override fun onTextModeClicked() {
        coroutineScope.launch {
            when (menuInteractor.onTextModeClicked()) {
                TextModeAction.ClearRectangles -> cvUseCase.clearAllRectangles()
                TextModeAction.None -> Unit
            }
        }
    }

    override fun onTryToFindText(text: String, locale: String) {
        coroutineScope.launch {
            menuInteractor.hideDialog()

            val screenSizes = videoUseCase.observeScreenSizes().value ?: return@launch
            val trimmed = text.trim()
            if (menuInteractor.observeMenuState().value is MenuState.Recording) {
                cvUseCase.snapshotSelectedRectangles()
            }
            val found = cvUseCase.findTextRectangles(
                serial = currentState.serial,
                text = trimmed,
                locale = locale,
                screenSizes = screenSizes,
            )

            if (!found) {
                uiCommandsFlow.tryEmit(StreamingUiCommand.ShowNetworkError)
                return@launch
            }
            menuInteractor.onTextSearchSuccess(text = trimmed, locale = locale)
        }
    }

    override fun onSaveClicked() {
        coroutineScope.launch {
            when (val action = menuInteractor.onSaveClicked(currentState.record)) {
                is SaveAction.SaveTemplate -> {
                    _stateFlow.update {
                        it.copy(record = it.record.copy(flags = action.flags))
                    }
                    val screenSizes = videoUseCase.observeScreenSizes().value ?: return@launch
                    cvUseCase.saveSelectedRectangle(
                        serial = currentState.serial,
                        screenSizes = screenSizes,
                    )
                    cvUseCase.nextCvMode(CVMode.NO_CV)
                }

                is SaveAction.SaveTextSelection -> {
                    _stateFlow.update {
                        it.copy(
                            record = it.record.copy(
                                text = action.text,
                                locale = action.locale,
                                flags = action.flags,
                            )
                        )
                    }
                    if (action.flags.textFlag() == 0) {
                        cvUseCase.clearSelectedRectangles()
                    }
                    cvUseCase.nextCvMode(CVMode.NO_CV)
                }

                is SaveAction.SaveTyping -> {
                    val pressEvent = currentState.record.stepEvents.extractPressEvent()
                    _stateFlow.update {
                        it.copy(
                            record = it.record.copy(
                                text = action.text,
                                flags = action.flags,
                                stepEvents = pressEvent,
                            )
                        )
                    }
                    cvUseCase.nextCvMode(CVMode.NO_CV)
                }

                SaveAction.SaveStep -> saveStep()
            }
        }
    }

    private suspend fun saveStep() {
        val record = currentState.record
        val success = scripterRepository.saveScriptStep(
            serial = currentState.serial,
            name = record.recordName,
            step = ScriptStep(
                events = record.stepEvents,
                flags = record.flags,
                text = record.text,
                locale = record.locale,
            ),
        )
        if (!success) return

        uiCommandsFlow.tryEmit(StreamingUiCommand.ShowStepSavedSnackbar)
        startRecordingTime = 0L
        cvUseCase.nextCvMode(CVMode.NO_CV)
        cvUseCase.clearAllRectangles()
        menuInteractor.onStepSaved()
        _stateFlow.update {
            it.copy(record = it.record.clearStep())
        }
    }

    override fun onCancelClicked() {
        coroutineScope.launch {
            val previousMenu = menuInteractor.observeMenuState().value
            val wasRecording = menuInteractor.onCancelClicked(currentState.record)
            cvUseCase.nextCvMode(CVMode.NO_CV)

            if (wasRecording) {
                cvUseCase.clearSelectedRectangles()
                _stateFlow.update { it.copy(record = StreamingState.Record()) }
                return@launch
            }

            if (previousMenu is MenuState.SelectingCV || previousMenu is MenuState.SelectingText) {
                cvUseCase.restoreSelectedRectangles()
            }
        }
    }

    override fun onSavedRecordName(name: String) {
        coroutineScope.launch {
            _stateFlow.update {
                it.copy(record = it.record.copy(recordName = name))
            }
            menuInteractor.onSavedRecordName()
            cvUseCase.nextCvMode(CVMode.NO_CV)
        }
    }

    override fun onSaveLocale(locale: String) {
        coroutineScope.launch {
            _stateFlow.update {
                it.copy(record = it.record.copy(locale = locale))
            }
            menuInteractor.onSaveLocale()
            getOrResetKeyboard()
        }
    }

    override fun onKeyboardInitClicked() {
        coroutineScope.launch {
            menuInteractor.setKeyboardLoadingState(true)

            val result = scripterRepository.resetKeyboard(
                serial = currentState.serial,
                locale = currentState.record.locale,
            )

            if (result is ApiResponse.Success) {
                setupKeyboardRects(result.data)
            }
        }
    }

    override fun onKeyboardEdited(addNew: Boolean) {
        coroutineScope.launch {
            val transition = menuInteractor.onKeyboardEdited(addNew)
            when (transition) {
                KeyboardEditTransition.ShowCvRectangles -> {
                    cvUseCase.clearSelectedRectangles()
                    cvUseCase.nextCvMode(CVMode.CV_RECTS)
                }

                KeyboardEditTransition.ShowKeyboardButtons -> {
                    cvUseCase.nextCvMode(CVMode.NO_CV)
                    cvUseCase.clearSelectedRectangles()
                }

                KeyboardEditTransition.None -> Unit
            }
        }
    }

    override fun onEditKeyboardButtonSaved(name: String) {
        coroutineScope.launch {
            val menuState = menuInteractor.observeMenuState().value
            val oldName = (menuState as? MenuState.Keyboard)?.oldKey.orEmpty()
            val screenSizes = videoUseCase.observeScreenSizes().value ?: return@launch
            val success = cvUseCase.editKeyboardSelectedRectangle(
                serial = currentState.serial,
                locale = currentState.record.locale,
                oldName = oldName,
                newName = name,
                screenSizes = screenSizes,
            )
            menuInteractor.hideDialog()
            cvUseCase.clearSelectedRectangles()
            if (success) {
                menuInteractor.setKeyboardLoadingState(true)
                getOrResetKeyboard()
                return@launch
            }
            uiCommandsFlow.tryEmit(StreamingUiCommand.ShowNetworkError)
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

        _stateFlow.update {
            it.copy(keyboard = it.keyboard.copy(buttons = buttons))
        }
        menuInteractor.setKeyboardLoadingState(false)
    }

    private fun showKeyboardError() {
        menuInteractor.setKeyboardLoadingState(false)
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
