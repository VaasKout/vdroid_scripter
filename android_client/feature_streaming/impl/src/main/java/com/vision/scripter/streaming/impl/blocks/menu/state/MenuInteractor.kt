package com.vision.scripter.streaming.impl.blocks.menu.state

import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.data.api.models.Parameter
import com.vision.scripter.streaming.impl.screen.main.state.CVMode
import com.vision.scripter.streaming.impl.screen.main.state.MenuState
import com.vision.scripter.streaming.impl.screen.main.state.SPACE_KEY
import com.vision.scripter.streaming.impl.screen.main.state.TEXT
import com.vision.scripter.streaming.impl.screen.main.state.TYPE_TEXT
import com.vision.scripter.streaming.impl.screen.main.state.increment
import com.vision.scripter.streaming.impl.screen.main.state.toType
import com.vision.scripter.streaming.impl.screen.main.state.toggleDetection
import com.vision.scripter.ui.CommandFlow
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.cancel
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import javax.inject.Inject

@ViewModelScoped
class MenuInteractor @Inject constructor(
    coroutineScopeFactory: CoroutineScopeFactory,
) : MenuUiStateHolder {

    private val coroutineScope: CoroutineScope =
        coroutineScopeFactory.createBackgroundScope("menu_interactor")

    override val uiCommandsFlow: CommandFlow<MenuUiCommand> =
        CommandFlow(coroutineScope)

    private val _menuState = MutableStateFlow<MenuState>(MenuState.Usual())
    fun observeMenuState(): StateFlow<MenuState> = _menuState.asStateFlow()

    private val _dialogState = MutableStateFlow(DialogState.NONE)
    fun observeDialogState(): StateFlow<DialogState> = _dialogState.asStateFlow()

    private val _events = MutableSharedFlow<MenuEvent>(extraBufferCapacity = 16)
    fun observeEvents(): SharedFlow<MenuEvent> = _events.asSharedFlow()

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

            else -> {}
        }
    }

    override fun onCvModeClicked() {
        when (val state = _menuState.value) {
            is MenuState.Recording -> {
                _menuState.update { MenuState.SelectingCV(cvMode = CVMode.CV_RECTS) }
                _events.tryEmit(MenuEvent.CvModeClicked(CVMode.CV_RECTS))
            }

            is MenuState.SelectingCV -> {
                val cvMode = state.cvMode.toggleDetection()
                _menuState.update { state.copy(cvMode = cvMode) }
                _events.tryEmit(MenuEvent.CvModeClicked(cvMode))
            }

            is MenuState.Usual -> {
                val newCvMode = state.localCvMode.increment()
                _menuState.update { state.copy(localCvMode = newCvMode) }
                _events.tryEmit(MenuEvent.CvModeClicked(newCvMode))
            }

            else -> {}
        }
    }

    override fun onTextModeClicked() {
        val state = _menuState.value
        if (state is MenuState.Usual && state.textHighlighted) {
            _menuState.update { state.copy(textHighlighted = false) }
            _events.tryEmit(MenuEvent.ClearRectangles)
            return
        }
        _dialogState.update { DialogState.TEXT }
    }

    override fun onTryToFindText(text: String, locale: String) {
        hideDialog()
        _events.tryEmit(MenuEvent.FindText(text = text.trim(), locale = locale))
    }

    fun onTextSearchSuccess(text: String, locale: String) {
        when (val state = _menuState.value) {
            is MenuState.Recording -> {
                _menuState.update {
                    MenuState.SelectingText(text = text, locale = locale)
                }
            }

            is MenuState.Usual -> _menuState.update {
                state.copy(localCvMode = CVMode.NO_CV, textHighlighted = true)
            }

            else -> {}
        }
    }

    override fun onKeyboardInitClicked() {
        setKeyboardLoadingState(true)
        _events.tryEmit(MenuEvent.KeyboardInit)
    }

    override fun onKeyboardEdited(addNew: Boolean) {
        val state = _menuState.value
        if (state !is MenuState.Keyboard) return
        val newState = when {
            !addNew -> state.copy(editing = !state.editing)
            else -> state.copy(showCvRectangles = !state.showCvRectangles)
        }
        _menuState.update { newState }
    }

    override fun onEditKeyboardButtonSaved(newKey: String) {
        val state = _menuState.value
        if (state !is MenuState.Keyboard) return
        hideDialog()
        _events.tryEmit(
            MenuEvent.EditKeyboardButton(
                newKey = newKey,
                oldKey = state.oldKey,
            )
        )
    }

    override fun onSaveClicked() {
        val param = when (val state = _menuState.value) {
            is MenuState.SelectingCV -> {
                _menuState.update { MenuState.Recording() }
                val key = state.cvMode.toType()
                Parameter(type = key, value = "", locale = "")
            }

            is MenuState.SelectingText -> {
                _menuState.update { MenuState.Recording() }
                Parameter(type = TEXT, value = state.text, locale = state.locale)
            }

            is MenuState.Keyboard -> {
                _menuState.update { MenuState.Recording() }
                Parameter(type = TYPE_TEXT, value = state.typeText)
            }

            else -> null
        }
        _events.tryEmit(MenuEvent.SaveClicked(param))
    }

    fun onScriptSaved() {
        _menuState.update { MenuState.Usual(expanded = true) }
    }

    override fun onCancelClicked() {
        when (val state = _menuState.value) {
            is MenuState.Recording -> {
                _menuState.update { MenuState.Usual(expanded = true) }
                _events.tryEmit(MenuEvent.RecordCancelled)
            }

            is MenuState.Keyboard -> {
                if (state.fromUsual) {
                    _menuState.update { MenuState.Usual(expanded = true) }
                    return
                }
                _menuState.update { MenuState.Recording() }
            }

            else -> {
                _menuState.update { MenuState.Recording() }
            }
        }
    }

    override fun onExitClicked() {
        uiCommandsFlow.tryEmit(MenuUiCommand.ExitCommand)
    }

    override fun onTimeoutSaved(timeout: Int) {
        updateTimeoutState(timeout > 0)
        hideDialog()
        val menuState = _menuState.value
        if (menuState is MenuState.Recording) {
            _menuState.update { menuState.copy(recordTimeout = timeout) }
        }
        _events.tryEmit(MenuEvent.TimeoutSaved(timeout))
    }

    override fun onSavedRecordName(name: String) {
        hideDialog()
        _menuState.update { MenuState.Recording() }
        _events.tryEmit(MenuEvent.SaveRecordName(name))
    }

    override fun onSaveLocale(locale: String) {
        hideDialog()
        val updated = when (_menuState.value) {
            is MenuState.Recording -> MenuState.Keyboard()
            is MenuState.Usual -> MenuState.Keyboard(fromUsual = true)
            else -> _menuState.value
        }
        _menuState.update { updated }
        _events.tryEmit(MenuEvent.SaveLocale(locale))
    }

    override fun onDialogDismissed() {
        hideDialog()
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

    private fun hideDialog() {
        _dialogState.update { DialogState.NONE }
    }

    fun clear() {
        coroutineScope.cancel()
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
