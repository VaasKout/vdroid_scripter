package com.vision.scripter.editscript.impl.state

import androidx.lifecycle.ViewModel
import com.vision.scripter.editscript.impl.ui.EditScriptUiStateHolder
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject

@HiltViewModel
class EditScriptViewModel @Inject constructor(
    val editScriptInteractor: EditScriptInteractor,
) : ViewModel(), EditScriptUiStateHolder by editScriptInteractor {

    override fun onCleared() {
        super.onCleared()
        editScriptInteractor.clear()
    }
}
