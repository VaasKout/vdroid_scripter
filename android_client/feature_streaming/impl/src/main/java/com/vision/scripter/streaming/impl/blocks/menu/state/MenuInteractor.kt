package com.vision.scripter.streaming.impl.blocks.menu.state

import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.data.api.models.TYPE_TEXT
import com.vision.scripter.streaming.impl.screen.main.state.CVMode
import com.vision.scripter.streaming.impl.screen.main.state.MenuState
import com.vision.scripter.streaming.impl.screen.main.state.SPACE_KEY
import com.vision.scripter.streaming.impl.screen.main.state.classFlag
import com.vision.scripter.streaming.impl.screen.main.state.combineDetection
import com.vision.scripter.streaming.impl.screen.main.state.combineText
import com.vision.scripter.streaming.impl.screen.main.state.increment
import com.vision.scripter.streaming.impl.screen.main.state.nextDetection
import com.vision.scripter.streaming.impl.screen.main.state.nextTemplateActive
import com.vision.scripter.streaming.impl.screen.main.state.nextText
import com.vision.scripter.streaming.impl.screen.main.state.templateFlag
import com.vision.scripter.streaming.impl.screen.main.state.textFlag
import com.vision.scripter.streaming.impl.screen.main.state.withFlag
import com.vision.scripter.streaming.impl.usecases.CvUseCase
import com.vision.scripter.ui.CommandFlow
import dagger.hilt.android.scopes.ActivityRetainedScoped
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@ActivityRetainedScoped
class MenuInteractor @Inject constructor(
    coroutineScopeFactory: CoroutineScopeFactory,
    private val cvUseCase: CvUseCase,
) : MenuUiStateHolder {

    private val coroutineScope: CoroutineScope =
        coroutineScopeFactory.createBackgroundScope("menu_interactor")

    override val uiCommandsFlow: CommandFlow<MenuUiCommand> =
        CommandFlow(coroutineScope)

    private val _menuState = MutableStateFlow<MenuState>(MenuState.Usual())
    fun observeMenuState(): StateFlow<MenuState> = _menuState.asStateFlow()

    private val _dialogState = MutableStateFlow(DialogState.NONE)
    fun observeDialogState(): StateFlow<DialogState> = _dialogState.asStateFlow()

    private val _events = MutableSharedFlow<MenuEvent>(replay = 1)
    fun observeEvents(): SharedFlow<MenuEvent> = _events.asSharedFlow()

    private var recordingFlags = 0

    override fun onScriptModeClicked() {
        _dialogState.update { DialogState.RECORD }
    }

    override fun onTimeoutClicked() {
        _dialogState.update { DialogState.TIMEOUT }
    }

    override fun onKeyboardClicked() {
        _dialogState.update { DialogState.KEYBOARD }
    }

    override fun onExpandClicked() {
        val state = _menuState.value
        if (state is MenuState.Usual) {
            _menuState.update { state.copy(expanded = !state.expanded) }
        }
    }

    override fun onRecordingClicked() {
        when (val state = _menuState.value) {
            is MenuState.Recording -> _menuState.update {
                state.copy(controlRecording = !state.controlRecording)
            }

            is MenuState.Keyboard -> _menuState.update {
                state.copy(
                    recordingKeyboard = !state.recordingKeyboard,
                    typeText = "",
                )
            }

            else -> Unit
        }
    }

    override fun onCvModeClicked() {
        coroutineScope.launch {
            when (val state = _menuState.value) {
                is MenuState.Recording -> {
                    val template = state.flags.templateFlag().nextTemplateActive()
                    recordingFlags = state.flags
                    _menuState.update { MenuState.SelectingCV(flags = template) }
                    cvUseCase.snapshotSelectedRectangles()
                    cvUseCase.nextCvMode(CVMode.CV_RECTS)
                }

                is MenuState.SelectingCV -> {
                    val detection = state.flags.nextDetection()
                    val cvMode = when {
                        detection.templateFlag() != 0 -> CVMode.CV_RECTS
                        detection.classFlag() != 0 -> CVMode.YOLO
                        else -> CVMode.NO_CV
                    }
                    val sourceChanged =
                        (state.flags.classFlag() != 0) != (detection.classFlag() != 0)
                    _menuState.update { state.copy(flags = detection) }
                    cvUseCase.nextCvMode(cvMode)
                    if (detection == 0 || sourceChanged) {
                        cvUseCase.clearSelectedRectangles()
                    }
                }

                is MenuState.Usual -> {
                    val newCvMode = state.cvMode.increment()
                    _menuState.update { state.copy(cvMode = newCvMode) }
                    cvUseCase.nextCvMode(newCvMode)
                }

                else -> Unit
            }
        }
    }

    override fun onTextModeClicked() {
        when (val state = _menuState.value) {
            is MenuState.Recording -> _dialogState.update { DialogState.TEXT }

            is MenuState.SelectingText -> _menuState.update {
                state.copy(flags = state.flags.nextText())
            }

            is MenuState.Usual -> {
                if (state.textHighlighted) {
                    _menuState.update { state.copy(textHighlighted = false) }
                    cvUseCase.clearAllRectangles()
                    return
                }
                _dialogState.update { DialogState.TEXT }
            }

            else -> Unit
        }
    }

    override fun onTryToFindText(text: String, locale: String) {
        coroutineScope.launch {
            hideDialog()
            if (_menuState.value is MenuState.Recording) {
                cvUseCase.snapshotSelectedRectangles()
            }
            _events.tryEmit(MenuEvent.FindText(text = text.trim(), locale = locale))
        }
    }

    fun onTextSearchSuccess(text: String, locale: String) {
        when (val state = _menuState.value) {
            is MenuState.Recording -> {
                recordingFlags = state.flags
                _menuState.update {
                    MenuState.SelectingText(
                        flags = state.flags.textFlag().nextText(),
                        text = text,
                        locale = locale,
                    )
                }
            }

            is MenuState.Usual -> _menuState.update {
                state.copy(
                    textHighlighted = true,
                    cvMode = CVMode.NO_CV,
                )
            }

            else -> Unit
        }
    }

    override fun onKeyboardInitClicked() {
        setKeyboardLoadingState(true)
        _events.tryEmit(MenuEvent.KeyboardInit)
    }

    override fun onKeyboardEdited(addNew: Boolean) {
        coroutineScope.launch {
            val state = _menuState.value
            if (state !is MenuState.Keyboard) return@launch
            val newState = when {
                !addNew -> state.copy(editing = !state.editing)
                else -> state.copy(showCvRectangles = !state.showCvRectangles)
            }
            _menuState.update { newState }

            if (newState.showCvRectangles) {
                cvUseCase.clearSelectedRectangles()
                cvUseCase.nextCvMode(CVMode.CV_RECTS)
            }
            if (newState.editing) {
                cvUseCase.nextCvMode(CVMode.NO_CV)
                cvUseCase.clearSelectedRectangles()
            }
        }
    }

    override fun onEditKeyboardButtonSaved(name: String) {
        _events.tryEmit(MenuEvent.EditKeyboardButton(name))
    }

    override fun onSaveClicked() {
        coroutineScope.launch {
            when (val state = _menuState.value) {
                is MenuState.SelectingCV -> {
                    val flags = recordingFlags.combineDetection(state.flags)
                    recordingFlags = flags
                    _menuState.update { MenuState.Recording(flags = flags) }
                    cvUseCase.nextCvMode(CVMode.NO_CV)
                    _events.tryEmit(MenuEvent.SaveTemplate(flags = flags))
                }

                is MenuState.SelectingText -> {
                    val flags = recordingFlags.combineText(state.flags)
                    recordingFlags = flags
                    _menuState.update { MenuState.Recording(flags = flags) }
                    if (flags.textFlag() == 0) {
                        cvUseCase.clearSelectedRectangles()
                    }
                    cvUseCase.nextCvMode(CVMode.NO_CV)
                    val hasText = state.flags != 0
                    _events.tryEmit(
                        MenuEvent.SaveText(
                            text = if (hasText) state.text else "",
                            locale = if (hasText) state.locale else "",
                            flags = flags,
                        )
                    )
                }

                is MenuState.Keyboard -> {
                    val flags = recordingFlags.withFlag(TYPE_TEXT, state.typeText.isNotEmpty())
                    recordingFlags = flags
                    _menuState.update { MenuState.Recording(flags = flags) }
                    cvUseCase.nextCvMode(CVMode.NO_CV)
                    _events.tryEmit(MenuEvent.SaveTyping(text = state.typeText, flags = flags))
                }

                else -> _events.tryEmit(MenuEvent.SaveStep)
            }
        }
    }

    suspend fun onStepSaved() {
        recordingFlags = 0
        _menuState.update { MenuState.Recording() }
        cvUseCase.nextCvMode(CVMode.NO_CV)
        cvUseCase.clearAllRectangles()
    }

    override fun onCancelClicked() {
        coroutineScope.launch {
            when (val state = _menuState.value) {
                is MenuState.Recording -> {
                    recordingFlags = 0
                    _menuState.update { MenuState.Usual(expanded = true) }
                    cvUseCase.nextCvMode(CVMode.NO_CV)
                    cvUseCase.clearSelectedRectangles()
                    _events.tryEmit(MenuEvent.RecordCancelled)
                }

                is MenuState.Keyboard -> {
                    cvUseCase.nextCvMode(CVMode.NO_CV)
                    if (state.fromUsual) {
                        recordingFlags = 0
                        _menuState.update { MenuState.Usual(expanded = true) }
                        return@launch
                    }
                    _menuState.update { MenuState.Recording(flags = recordingFlags) }
                }

                is MenuState.SelectingCV -> {
                    _menuState.update { MenuState.Recording(flags = recordingFlags) }
                    cvUseCase.nextCvMode(CVMode.NO_CV)
                    cvUseCase.restoreSelectedRectangles()
                }

                is MenuState.SelectingText -> {
                    _menuState.update { MenuState.Recording(flags = recordingFlags) }
                    cvUseCase.nextCvMode(CVMode.NO_CV)
                    cvUseCase.restoreSelectedRectangles()
                }

                else -> Unit
            }
        }
    }

    override fun onExitClicked() {
        uiCommandsFlow.tryEmit(MenuUiCommand.ExitCommand)
    }

    override fun onTimeoutSaved(timeout: Int) {
        updateTimeoutState(timeout > 0)
        hideDialog()
        _events.tryEmit(MenuEvent.TimeoutSaved(timeout))
    }

    override fun onSavedRecordName(name: String) {
        coroutineScope.launch {
            hideDialog()
            recordingFlags = 0
            _menuState.update { MenuState.Recording() }
            cvUseCase.nextCvMode(CVMode.NO_CV)
            _events.tryEmit(MenuEvent.SaveRecordName(name))
        }
    }

    override fun onSaveLocale(locale: String) {
        hideDialog()
        val updated = when (val state = _menuState.value) {
            is MenuState.Recording -> {
                recordingFlags = state.flags
                MenuState.Keyboard()
            }

            is MenuState.Usual -> {
                recordingFlags = 0
                MenuState.Keyboard(fromUsual = true)
            }

            else -> state
        }
        _menuState.update { updated }
        _events.tryEmit(MenuEvent.SaveLocale(locale))
    }

    override fun onDialogDismissed() {
        if (_menuState.value is MenuState.Keyboard) {
            cvUseCase.clearSelectedRectangles()
        }
        hideDialog()
        if (_menuState.value is MenuState.Recording) {
            _menuState.update { MenuState.Recording(flags = recordingFlags) }
        }
    }

    fun setKeyboardLoadingState(isLoading: Boolean) {
        val state = _menuState.value
        if (state is MenuState.Keyboard) {
            _menuState.update { state.copy(isLoadingKeyboard = isLoading) }
        }
    }

    fun appendTypedLetter(letter: String) {
        val state = _menuState.value
        if (state !is MenuState.Keyboard) return
        val updatedTypeText = buildString {
            append(state.typeText)
            if (letter == SPACE_KEY) append(" ")
            else append(letter)
        }
        _menuState.update { state.copy(typeText = updatedTypeText) }
    }

    fun setKeyboardOldKey(oldKey: String) {
        val state = _menuState.value
        if (state !is MenuState.Keyboard) return
        _menuState.update {
            state.copy(oldKey = oldKey)
        }
    }

    private fun updateTimeoutState(customTimeout: Boolean) {
        val menuState = _menuState.value
        if (menuState is MenuState.Recording) {
            _menuState.update {
                menuState.copy(customTimeout = customTimeout)
            }
        }
    }

    fun onEditKeyboardRectangleSelected() {
        _dialogState.update { DialogState.EDIT_KEYBOARD }
    }

    fun hideDialog() {
        _dialogState.update { DialogState.NONE }
    }
}

enum class DialogState {
    NONE,
    RECORD,
    TEXT,
    KEYBOARD,
    EDIT_KEYBOARD,
    TIMEOUT;
}
