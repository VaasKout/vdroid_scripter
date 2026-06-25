package com.vision.scripter.streaming.impl.state

import com.vision.scripter.data.api.models.TYPE_TEXT
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import javax.inject.Inject

@ViewModelScoped
class MenuInteractor @Inject constructor() {

    private val _menuState = MutableStateFlow<MenuState>(MenuState.Usual())
    fun observeMenuState(): StateFlow<MenuState> = _menuState.asStateFlow()

    private val _dialogState = MutableStateFlow(DialogState.NONE)
    fun observeDialogState(): StateFlow<DialogState> = _dialogState.asStateFlow()

    fun onScriptModeClicked() {
        _dialogState.update { DialogState.RECORD }
    }

    fun onTimeoutClicked() {
        _dialogState.update { DialogState.TIMEOUT }
    }

    fun onExpandClicked() {
        val state = _menuState.value
        if (state is MenuState.Usual) {
            _menuState.update { state.copy(expanded = !state.expanded) }
        }
    }

    fun onRecordingClicked() {
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

    fun onKeyboardClicked() {
        _dialogState.update { DialogState.KEYBOARD }
    }

    fun onCvModeClicked(): CvModeAction? {
        return when (val state = _menuState.value) {
            is MenuState.Recording -> {
                val template = state.flags.templateFlag().nextTemplateActive()
                _menuState.update { MenuState.SelectingCV(flags = template) }
                CvModeAction(newCvMode = CVMode.CV_RECTS)
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
                CvModeAction(
                    newCvMode = cvMode,
                    disableSelection = detection == 0 || sourceChanged,
                )
            }

            is MenuState.Usual -> {
                val newCVMode = state.cvMode.increment()
                _menuState.update { state.copy(cvMode = newCVMode) }
                CvModeAction(newCvMode = newCVMode)
            }

            else -> null
        }
    }

    fun onTextModeClicked(): TextModeAction {
        return when (val state = _menuState.value) {
            is MenuState.Recording -> {
                _dialogState.update { DialogState.TEXT }
                TextModeAction.None
            }

            is MenuState.SelectingText -> {
                val text = state.flags.nextText()
                _menuState.update { state.copy(flags = text) }
                TextModeAction.None
            }

            is MenuState.Usual -> {
                if (state.textHighlighted) {
                    _menuState.update { state.copy(textHighlighted = false) }
                    return TextModeAction.ClearRectangles
                }
                _dialogState.update { DialogState.TEXT }
                TextModeAction.None
            }

            else -> TextModeAction.None
        }
    }

    fun onTextSearchSuccess(text: String, locale: String) {
        when (val state = _menuState.value) {
            is MenuState.Recording -> _menuState.update {
                MenuState.SelectingText(
                    flags = state.flags.textFlag().nextText(),
                    text = text,
                    locale = locale,
                )
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

    fun onSaveClicked(record: StreamingState.Record): SaveAction {
        return when (val state = _menuState.value) {
            is MenuState.SelectingCV -> {
                val flags = record.flags.combineDetection(state.flags)
                _menuState.update { MenuState.Recording(flags = flags) }
                SaveAction.SaveTemplate(flags = flags)
            }

            is MenuState.SelectingText -> {
                val flags = record.flags.combineText(state.flags)
                _menuState.update { MenuState.Recording(flags = flags) }
                val hasText = state.flags != 0
                SaveAction.SaveTextSelection(
                    text = if (hasText) state.text else "",
                    locale = if (hasText) state.locale else "",
                    flags = flags,
                )
            }

            is MenuState.Keyboard -> {
                val flags = record.flags.withFlag(TYPE_TEXT, state.typeText.isNotEmpty())
                _menuState.update { MenuState.Recording(flags = flags) }
                SaveAction.SaveTyping(text = state.typeText, flags = flags)
            }

            else -> SaveAction.SaveStep
        }
    }

    fun onStepSaved() {
        _menuState.update { MenuState.Recording() }
    }

    fun onCancelClicked(record: StreamingState.Record): Boolean {
        val state = _menuState.value
        val wasRecording = state is MenuState.Recording
        val backToUsual = wasRecording ||
                (state is MenuState.Keyboard && state.fromUsual)

        if (backToUsual) {
            _menuState.update { MenuState.Usual(expanded = true) }
        } else {
            _menuState.update { MenuState.Recording(flags = record.flags) }
        }
        return wasRecording
    }

    fun onSavedRecordName() {
        hideDialog()
        _menuState.update { MenuState.Recording() }
    }

    fun onSaveLocale() {
        hideDialog()
        val updated = when (val state = _menuState.value) {
            is MenuState.Recording -> MenuState.Keyboard()
            is MenuState.Usual -> MenuState.Keyboard(fromUsual = true)
            else -> state
        }
        _menuState.update { updated }
    }

    fun onDialogDismissed(record: StreamingState.Record) {
        val state = _menuState.value
        hideDialog()
        if (state is MenuState.Recording) {
            _menuState.update { MenuState.Recording(flags = record.flags) }
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

    fun onKeyboardEdited(addNew: Boolean): KeyboardEditTransition {
        val state = _menuState.value
        if (state !is MenuState.Keyboard) return KeyboardEditTransition.None
        val newState = when {
            !addNew -> state.copy(editing = !state.editing)
            else -> state.copy(showCvRectangles = !state.showCvRectangles)
        }
        _menuState.update { newState }
        return when {
            newState.showCvRectangles -> KeyboardEditTransition.ShowCvRectangles
            newState.editing -> KeyboardEditTransition.ShowKeyboardButtons
            else -> KeyboardEditTransition.None
        }
    }

    fun setKeyboardOldKey(oldKey: String) {
        val state = _menuState.value
        if (state !is MenuState.Keyboard) return
        _menuState.update {
            state.copy(oldKey = oldKey)
        }
    }

    fun updateTimeoutState(customTimeout: Boolean) {
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

data class CvModeAction(
    val newCvMode: CVMode,
    val disableSelection: Boolean = false,
)

sealed interface TextModeAction {
    data object None : TextModeAction
    data object ClearRectangles : TextModeAction
}

sealed interface SaveAction {
    data class SaveTemplate(val flags: Int) : SaveAction
    data class SaveTextSelection(
        val text: String,
        val locale: String,
        val flags: Int,
    ) : SaveAction

    data class SaveTyping(val text: String, val flags: Int) : SaveAction
    data object SaveStep : SaveAction
}

enum class DialogState {
    NONE,
    RECORD,
    TEXT,
    KEYBOARD,
    EDIT_KEYBOARD,
    TIMEOUT;
}

sealed interface KeyboardEditTransition {
    data object None : KeyboardEditTransition
    data object ShowKeyboardButtons : KeyboardEditTransition
    data object ShowCvRectangles : KeyboardEditTransition
}
