{-# LANGUAGE TypeSynonymInstances #-}

module Main (main) where

import MonadUtils

import GHC
import Outputable
import GHC.Paths ( libdir )
import DynFlags ( defaultDynFlags )

import Type
import HscTypes ( cm_binds, CoreModule(..), mkTypeEnv, HscEnv(..) )
import CoreSyn
import IdInfo
import Var as V
import Name as N
import OccName as O
import DriverPipeline as DP
import DriverPhases as DPh

import PrelNames as PN
import TysPrim as TP
import TysWiredIn as TW

import Unique as U
import SrcLoc as SL
import FastString as FS

main =
  defaultErrorHandler defaultDynFlags $ do
    runGhc (Just libdir) $ do
      dflags <- getSessionDynFlags
      setSessionDynFlags (setVerbosity 3 (setHscOut "out.s" (setOutputDir "." dflags)))
      dflags2 <- getSessionDynFlags

      ioTyCon   <- lookupTyCon ioTyConName
      runMainIO <- lookupId PN.runMainIOName
      compileCoreToObj False $ coreMod ioTyCon runMainIO

      hsc_env <- getSession
      DP.oneShot hsc_env DPh.StopLn [("out.s", Just DPh.As)] 

lookupTyCon name = do
  tyCon <- lookupName name
  case tyCon of
    Just (ATyCon tc) -> return tc
    _                -> error "No name found"

lookupId name = do
  tyThing <- lookupName name
  case tyThing of
    Just (AnId id)  -> return id
    _               -> error "No name found"


coreMod ioTyCon runMainIO = CoreModule PN.mAIN (mkTypeEnv []) [mainNR ioTyCon, runMainIONR ioTyCon runMainIO] []

mainNR ioTyCon = NonRec (mainVar ioTyCon) (App (Var $ putStrLnVar ioTyCon) (App (Var unpackCStringVar) (mkStringLit "Hello")))

runMainIONR ioTyCon runMainIO = NonRec (rootMainVar ioTyCon) (App (App (Var runMainIO) (Type TW.unitTy)) (Var $ mainVar ioTyCon))

mainVar ioTyCon = globaliseId $ mkExportedLocalVar VanillaId (mkExtName PN.mAIN "main" 1) (mkTyConApp ioTyCon [TW.unitTy]) vanillaIdInfo -- Vanilla??

rootMainVar ioTyCon = globaliseId $ mkExportedLocalVar VanillaId 
                                      (N.mkExternalName PN.rootMainKey PN.rOOT_MAIN (mkVarOccFS (fsLit "main")) mkNoSrcSpan) 
                                      (mkTyConApp ioTyCon [TW.unitTy]) vanillaIdInfo

putStrLnVar ioTyCon = globaliseId $ 
  mkExportedLocalVar VanillaId (mkExtName PN.sYSTEM_IO "putStrLn" 2) (mkFunTy TW.stringTy $ mkTyConApp ioTyCon [TW.unitTy]) vanillaIdInfo

unpackCStringVar = globaliseId $ mkExportedLocalVar VanillaId PN.unpackCStringName (mkFunTy TP.addrPrimTy TW.stringTy) vanillaIdInfo

-- TODO use UniqSupply
mkExtName mod def uniq = N.mkExternalName (U.mkPseudoUniqueH uniq) mod (O.mkVarOccFS $ FS.fsLit def) mkNoSrcSpan  

mkNoSrcSpan = SL.mkSrcSpan (SL.mkSrcLoc (FS.fsLit "<prog>") 0 0) (SL.mkSrcLoc (FS.fsLit "<prog>") 0 0)

setObjectDir  f d = d{ objectDir  = Just f}
setHiDir      f d = d{ hiDir      = Just f}
setStubDir    f d = d{ stubDir    = Just f, includePaths = f : includePaths d }
setOutputDir  f = setObjectDir f . setHiDir f . setStubDir f
setHscOut     f d = d{ hscOutName = f }
setVerbosity  v d = d{ verbosity = v }

