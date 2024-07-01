@echo off

REM --------------------------------------------------------------------------
REM     Program:  sample.bat
REM  Programmer:  George Kerber
REM     Written:  02/14/95
REM     Revised:  02/14/95
REM --------------------------------------------------------------------------


REM put up a press box using data in the sample.dat file
press -z -i sample.dat

if ERRORLEVEL 1 goto cancel
REM put up a press box using -l "text" command line data
press -s black -b green -l "You selected %%bOKAY!%%b" -l " " -l "View the sample.bat file to see how easy this is..."
goto done


:cancel

REM put up a press box using -l "text" command line data
press -g -m -b red -l "You selected %%bCANCEL!%%b" -l " " -l "View the sample.bat file to see how easy this is..."
goto done

:done
press -t 2 -b cyan -s black -l "All Done" -l "(timeout in 2 seconds)"
