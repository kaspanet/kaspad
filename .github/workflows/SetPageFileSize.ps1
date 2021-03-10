<#
# MIT License (MIT) Copyright (c) 2020 Maxim Lobanov and contributors
# Source: https://github.com/al-cheb/configure-pagefile-action/blob/master/scripts/SetPageFileSize.ps1
.SYNOPSIS
  Configure Pagefile on Windows machine
.NOTES
  Author:         Aleksandr Chebotov

.EXAMPLE
  SetPageFileSize.ps1 -MinimumSize 4GB -MaximumSize 8GB -DiskRoot "D:"
#>

param(
    [System.UInt64] $MinimumSize = 8gb ,
    [System.UInt64] $MaximumSize = 8gb ,
    [System.String] $DiskRoot = "D:"
)

# https://referencesource.microsoft.com/#System.IdentityModel/System/IdentityModel/NativeMethods.cs,619688d876febbe1
# https://www.geoffchappell.com/studies/windows/km/ntoskrnl/api/mm/modwrite/create.htm
# https://referencesource.microsoft.com/#mscorlib/microsoft/win32/safehandles/safefilehandle.cs,9b08210f3be75520
# https://referencesource.microsoft.com/#mscorlib/system/security/principal/tokenaccesslevels.cs,6eda91f498a38586
# https://www.autoitscript.com/forum/topic/117993-api-ntcreatepagingfile/

$source = @'
using System;
using System.ComponentModel;
using System.Diagnostics;
using System.Runtime.InteropServices;
using System.Security.Principal;
using System.Text;
using Microsoft.Win32;
using Microsoft.Win32.SafeHandles;

namespace Util
{
    class NativeMethods
    {
        [StructLayout(LayoutKind.Sequential)]
        internal struct LUID
        {
            internal uint LowPart;
            internal uint HighPart;
        }
    
        [StructLayout(LayoutKind.Sequential)]
        internal struct LUID_AND_ATTRIBUTES
        {
            internal LUID Luid;
            internal uint Attributes;
        }
    
        [StructLayout(LayoutKind.Sequential)]
        internal struct TOKEN_PRIVILEGE
        {
            internal uint PrivilegeCount;
            internal LUID_AND_ATTRIBUTES Privilege;
    
            internal static readonly uint Size = (uint)Marshal.SizeOf(typeof(TOKEN_PRIVILEGE));
        }

        [StructLayoutAttribute(LayoutKind.Sequential, CharSet = CharSet.Unicode)]
        internal struct UNICODE_STRING
        {
            internal UInt16 length;
            internal UInt16 maximumLength;
            internal string buffer;
        }

        [DllImport("kernel32.dll", SetLastError=true)]
        internal static extern IntPtr LocalFree(IntPtr handle);

        [DllImport("advapi32.dll", ExactSpelling = true, CharSet = CharSet.Unicode, SetLastError = true, PreserveSig = false)]
        internal static extern bool LookupPrivilegeValueW(
            [In] string lpSystemName,
            [In] string lpName,
            [Out] out LUID luid
        );

        [DllImport("advapi32.dll", SetLastError = true, PreserveSig = false)]
        internal static extern bool AdjustTokenPrivileges(
            [In] SafeCloseHandle tokenHandle,
            [In] bool disableAllPrivileges,
            [In] ref TOKEN_PRIVILEGE newState,
            [In] uint bufferLength,
            [Out] out TOKEN_PRIVILEGE previousState,
            [Out] out uint returnLength
        );

        [DllImport("advapi32.dll", CharSet = CharSet.Auto, SetLastError = true, PreserveSig = false)]
        internal static extern bool OpenProcessToken(
            [In] IntPtr processToken,
            [In] int desiredAccess,
            [Out] out SafeCloseHandle tokenHandle
        );

        [DllImport("ntdll.dll", CharSet = CharSet.Unicode, SetLastError = true, CallingConvention = CallingConvention.StdCall)]
        internal static extern Int32 NtCreatePagingFile(
            [In] ref UNICODE_STRING pageFileName, 
            [In] ref Int64 minimumSize, 
            [In] ref Int64 maximumSize, 
            [In] UInt32 flags
        );

        [DllImport("kernel32.dll", CharSet = CharSet.Unicode, SetLastError = true)]
        internal static extern uint QueryDosDeviceW(
            string lpDeviceName,
            StringBuilder lpTargetPath,
            int ucchMax
        );
    }

    public sealed class SafeCloseHandle: SafeHandleZeroOrMinusOneIsInvalid 
    {
        [DllImport("kernel32.dll", ExactSpelling = true, SetLastError = true)]
        internal extern static bool CloseHandle(IntPtr handle);

        private SafeCloseHandle() : base(true)
        {
        }
 
        public SafeCloseHandle(IntPtr preexistingHandle, bool ownsHandle) : base(ownsHandle) 
        {
            SetHandle(preexistingHandle);
        }

        override protected bool ReleaseHandle()
        {
            return CloseHandle(handle);
        }
    }

    public class PageFile
    {
        public static void SetPageFileSize(long minimumValue, long maximumValue, string lpDeviceName)
        {
            SetPageFilePrivilege();
            StringBuilder lpTargetPath = new StringBuilder(260);

            UInt32 resultQueryDosDevice = NativeMethods.QueryDosDeviceW(lpDeviceName, lpTargetPath, lpTargetPath.Capacity);
            if (resultQueryDosDevice == 0)
            {
                throw new Win32Exception(Marshal.GetLastWin32Error());
            }

            string pageFilePath = lpTargetPath.ToString() + "\\pagefile.sys";

            NativeMethods.UNICODE_STRING pageFileName = new NativeMethods.UNICODE_STRING
            {
                length = (ushort)(pageFilePath.Length * 2),
                maximumLength = (ushort)(2 * (pageFilePath.Length + 1)),
                buffer = pageFilePath
            };

            Int32 resultNtCreatePagingFile = NativeMethods.NtCreatePagingFile(ref pageFileName, ref minimumValue, ref maximumValue, 0);
            if (resultNtCreatePagingFile != 0)
            {
                throw new Win32Exception(Marshal.GetLastWin32Error());
            }

            Console.WriteLine("PageFile: {0} / {1} bytes for {2}", minimumValue, maximumValue, pageFilePath);
        }

        static void SetPageFilePrivilege()
        {
            const int SE_PRIVILEGE_ENABLED = 0x00000002;
            const int AdjustPrivileges = 0x00000020;
            const int Query = 0x00000008;

            NativeMethods.LUID luid;
            NativeMethods.LookupPrivilegeValueW(null, "SeCreatePagefilePrivilege", out luid);

            SafeCloseHandle hToken;
            NativeMethods.OpenProcessToken(
                Process.GetCurrentProcess().Handle,
                AdjustPrivileges | Query,
                out hToken
            );

            NativeMethods.TOKEN_PRIVILEGE previousState;
            NativeMethods.TOKEN_PRIVILEGE newState;
            uint previousSize = 0;
            newState.PrivilegeCount = 1;
            newState.Privilege.Luid = luid;
            newState.Privilege.Attributes = SE_PRIVILEGE_ENABLED;

            NativeMethods.AdjustTokenPrivileges(hToken, false, ref newState, NativeMethods.TOKEN_PRIVILEGE.Size, out previousState, out previousSize);
        }
    }
}
'@

Add-Type -TypeDefinition $source

# Set SetPageFileSize
[Util.PageFile]::SetPageFileSize($minimumSize, $maximumSize, $diskRoot)