package com.vision.scripter.streaming.impl.data

import com.vision.scripter.data.api.CvStreamer
import com.vision.scripter.data.api.ScripterDataSource
import com.vision.scripter.data.api.models.CvRectangle
import dagger.hilt.android.scopes.ViewModelScoped
import javax.inject.Inject

@ViewModelScoped
class KeyboardRepository @Inject constructor(
    private val cvStreamer: CvStreamer,
    private val scripterDataSource: ScripterDataSource,
) {
    suspend fun editKeyboardSelectedRectangle(
        serial: String,
        locale: String,
        oldName: String,
        newName: String,
        rectangle: CvRectangle,
    ): Boolean {
        if (newName.isEmpty()) return false

        if (oldName.isNotEmpty()) {
            val deleted = scripterDataSource.deleteButton(
                serial = serial,
                locale = locale,
                name = oldName
            )
            if (!deleted) return false
        }

        return scripterDataSource.editKeyboard(
            serial = serial,
            locale = locale,
            name = newName,
            rectangle = rectangle,
        )
    }

}